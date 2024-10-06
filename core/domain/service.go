package domain

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/salt/log"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/robfig/cron/v3"
)

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
}

type OrgService interface {
	ListByUser(ctx context.Context, principal authenticate.Principal, filter organization.Filter) ([]organization.Organization, error)
	AddMember(ctx context.Context, orgID, relationName string, principal authenticate.Principal) error
	Get(ctx context.Context, id string) (organization.Organization, error)
}

type Service struct {
	repository  Repository
	userService UserService
	orgService  OrgService
	cron        *cron.Cron
	log         log.Logger
}

const (
	DNSChallenge       = "_frontier-domain-verification=%s"
	txtLength          = 40
	DefaultTokenExpiry = time.Hour * 24 * 7 // 7 days
	refreshTime        = "0 0 * * *"        // Once a day at midnight (UTC)
)

func NewService(logger log.Logger, repository Repository, userService UserService, orgService OrgService) *Service {
	return &Service{
		repository:  repository,
		userService: userService,
		orgService:  orgService,
		cron:        cron.New(),
		log:         logger,
	}
}

// Get an organization's whitelisted domain from the database
func (s Service) Get(ctx context.Context, id string) (Domain, error) {
	return s.repository.Get(ctx, id)
}

// List all whitelisted domains for an organization (filter by verified boolean)
func (s Service) List(ctx context.Context, flt Filter) ([]Domain, error) {
	return s.repository.List(ctx, flt)
}

// Remove an organization's whitelisted domain from the database
func (s Service) Delete(ctx context.Context, id string) error {
	return s.repository.Delete(ctx, id)
}

// Creates a record for the domain in the database and returns the TXT record that needs to be added to the DNS for the domain verification
func (s Service) Create(ctx context.Context, domain Domain) (Domain, error) {
	orgResp, err := s.orgService.Get(ctx, domain.OrgID)
	if err != nil {
		return Domain{}, err
	}

	txtRecord, err := generateRandomTXT()
	if err != nil {
		return Domain{}, err
	}

	domain.OrgID = orgResp.ID // in case the orgName is provided in the request, replace with the orgID
	domain.Token = fmt.Sprintf(DNSChallenge, txtRecord)
	var domainResp Domain
	if domainResp, err = s.repository.Create(ctx, domain); err != nil {
		return Domain{}, err
	}

	return domainResp, nil
}

// VerifyDomain checks if the TXT record for the domain matches the token generated by Frontier for the domain verification
func (s Service) VerifyDomain(ctx context.Context, id string) (Domain, error) {
	domain, err := s.repository.Get(ctx, id)
	if err != nil {
		return Domain{}, ErrNotExist
	}

	txtRecords, err := net.LookupTXT(domain.Name)
	if err != nil {
		if strings.Contains(err.Error(), "no such host") {
			return Domain{}, ErrInvalidDomain
		}
		return Domain{}, err
	}

	for _, txtRecord := range txtRecords {
		if strings.TrimSpace(txtRecord) == strings.TrimSpace(domain.Token) {
			domain.State = Verified
			domain, err = s.repository.Update(ctx, domain)
			if err != nil {
				return Domain{}, err
			}
			return domain, nil
		}
	}

	return domain, ErrTXTrecordNotFound
}

// Join an organization as a member if the user domain matches the org whitelisted domains
func (s Service) Join(ctx context.Context, orgID string, userId string) error {
	orgResp, err := s.orgService.Get(ctx, orgID)
	if err != nil {
		return err
	}

	currUser, err := s.userService.GetByID(ctx, userId)
	if err != nil {
		return err
	}

	// check if user is already a member of the organization. if yes, do nothing and return nil
	userOrgs, err := s.orgService.ListByUser(ctx, authenticate.Principal{
		ID:   currUser.ID,
		Type: schema.UserPrincipal,
	}, organization.Filter{})
	if err != nil {
		return err
	}

	for _, org := range userOrgs {
		if org.ID == orgResp.ID {
			return nil
		}
	}

	userDomain := utils.ExtractDomainFromEmail(currUser.Email)
	if userDomain == "" {
		return user.ErrInvalidEmail
	}

	// check if user domain matches the org whitelisted domains
	orgTrustedDomains, err := s.List(ctx, Filter{
		OrgID: orgResp.ID,
		State: Verified,
	})
	if err != nil {
		return err
	}

	for _, dmn := range orgTrustedDomains {
		if userDomain == dmn.Name {
			if err = s.orgService.AddMember(ctx, orgResp.ID, schema.MemberRelationName, authenticate.Principal{
				ID:   currUser.ID,
				Type: schema.UserPrincipal,
			}); err != nil {
				return err
			}
			return nil
		}
	}

	return ErrDomainsMisMatch
}

func (s Service) ListJoinableOrgsByDomain(ctx context.Context, email string) ([]string, error) {
	domain := utils.ExtractDomainFromEmail(email)
	domains, err := s.repository.List(ctx, Filter{
		Name:  domain,
		State: Verified,
	})
	if err != nil {
		return nil, err
	}

	// check if user is already a member of the organization. if yes, do not include the org in the response
	currUser, err := s.userService.GetByID(ctx, email)
	if err != nil {
		return nil, err
	}

	userOrgs, err := s.orgService.ListByUser(ctx, authenticate.Principal{
		ID:   currUser.ID,
		Type: schema.UserPrincipal,
	}, organization.Filter{})
	if err != nil {
		return nil, err
	}

	var orgIDs []string
	var alreadyMember bool
	for _, domain := range domains {
		alreadyMember = false
		for _, org := range userOrgs {
			if org.ID == domain.OrgID {
				alreadyMember = true
				break
			}
		}
		if !alreadyMember {
			orgIDs = append(orgIDs, domain.OrgID)
		}
	}
	return orgIDs, nil
}

// InitDomainVerification starts a cron job that runs once a day to delete expired domain requests which are still in pending state after 7 days of creation
func (s Service) InitDomainVerification(ctx context.Context) error {
	_, err := s.cron.AddFunc(refreshTime, func() {
		if err := s.repository.DeleteExpiredDomainRequests(ctx); err != nil {
			s.log.Warn("error deleting expired domain requests", "err", err)
		}
	})
	if err != nil {
		return err
	}
	s.cron.Start()
	return nil
}

func (s Service) Close() error {
	return s.cron.Stop().Err()
}

func generateRandomTXT() (string, error) {
	randomBytes := make([]byte, txtLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode the random bytes in Base64
	txtRecord := base64.StdEncoding.EncodeToString(randomBytes)
	return txtRecord, nil
}
