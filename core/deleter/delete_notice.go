package deleter

import (
	"bytes"
	"context"
	"fmt"
	htmltemplate "html/template"
	"log/slog"
	"strconv"
	"strings"
	texttemplate "text/template"

	"github.com/raystack/frontier/billing/credit"

	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/membership"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"gopkg.in/mail.v2"
)

// plain fallbacks used when the config leaves the templates empty
const (
	defaultDeleteNoticeSubject = "Your organization was deleted"
	defaultDeleteNoticeBody    = `{{if .User.Title}}Hi {{.User.Title}},{{else}}Hi,{{end}}<br><br>Your organization <b>{{if .Org.Title}}{{.Org.Title}}{{else}}{{.Org.Name}}{{end}}</b> was deleted{{if .DeletedBy}} by <b>{{.DeletedBy}}</b>{{end}}.{{if .Purchased}} Unused purchased tokens will be settled by the support team.{{end}}`
)

type deleteNoticeData struct {
	// Amount is the total number of tokens the delete forfeited.
	Amount int64
	// User is the owner receiving this mail.
	User user.User
	// Org is the deleted organization.
	Org organization.Organization
	// Purchased is the share of Amount that came from purchases; only this
	// part is transferable.
	Purchased int64
	// DeletedBy identifies who ran the delete; empty when the caller is
	// not known.
	DeletedBy string
}

// accountTokens is what one billing account holds at delete time.
type accountTokens struct {
	Balance   int64
	Purchased int64
}

// deleteNotice is everything sendDeleteNotices needs once the org is gone.
// It has to be collected before teardown removes the owners and the token
// balances. Accounts keeps the per-account numbers so the teardown can audit
// them without reading the balances a second time.
type deleteNotice struct {
	Amount    int64
	Purchased int64
	Accounts  map[string]accountTokens
	// Balances holds every account's balance, positive or not, so the
	// blocker check can reuse the reads
	Balances map[string]int64
	Owners   []user.User
}

// collectDeleteNotice reads the token balances the delete is about to
// forfeit. It only reads; a failure here aborts the delete before anything
// is torn down.
//
// The amount is the whole remaining balance. Purchased is the share of it
// that came from purchases (source system.buy), with complimentary tokens
// (plan starter grants and awards) counted as spent first. Only the
// purchased share is settled with the customer.
func (d Service) collectDeleteNotice(ctx context.Context, customers []customer.Customer) (deleteNotice, error) {
	var total, purchased int64
	accounts := make(map[string]accountTokens, len(customers))
	balances := make(map[string]int64, len(customers))
	for _, c := range customers {
		balance, err := d.creditService.GetBalance(ctx, c.ID)
		if err != nil {
			return deleteNotice{}, fmt.Errorf("failed to check token balance of billing account[%s]: %w", c.ID, err)
		}
		balances[c.ID] = balance
		if balance > 0 {
			bought, err := d.purchasedTokens(ctx, c.ID, balance)
			if err != nil {
				return deleteNotice{}, err
			}
			total += balance
			purchased += bought
			accounts[c.ID] = accountTokens{Balance: balance, Purchased: bought}
		}
	}
	if total == 0 {
		return deleteNotice{Accounts: accounts, Balances: balances}, nil
	}

	return deleteNotice{
		Amount:    total,
		Purchased: purchased,
		Accounts:  accounts,
		Balances:  balances,
	}, nil
}

// resolveOwners finds the users holding the org owner role; every one of
// them gets the delete notice. It is best-effort: the notice email must not
// make the delete depend on the policy machinery, so a failed lookup logs
// and returns no owners.
func (d Service) resolveOwners(ctx context.Context, orgID string) []user.User {
	ownerRole, err := d.roleService.Get(ctx, schema.RoleOrganizationOwner)
	if err != nil {
		slog.WarnContext(ctx, "failed to resolve the organization owner role for the delete notice", "org_id", orgID, "error", err)
		return nil
	}
	members, err := d.membershipService.ListPrincipalsByResource(ctx, orgID, schema.OrganizationNamespace, membership.MemberFilter{
		PrincipalType: schema.UserPrincipal,
		RoleIDs:       []string{ownerRole.ID},
	})
	if err != nil {
		slog.WarnContext(ctx, "failed to list the organization owners for the delete notice", "org_id", orgID, "error", err)
		return nil
	}
	ownerIDs := make([]string, 0, len(members))
	for _, m := range members {
		ownerIDs = append(ownerIDs, m.PrincipalID)
	}
	owners, err := d.userService.GetByIDs(ctx, ownerIDs)
	if err != nil {
		slog.WarnContext(ctx, "failed to fetch the organization owners for the delete notice", "org_id", orgID, "error", err)
		return nil
	}
	return owners
}

// recordNoticeRecipients writes the owner ids to an audit record before
// teardown starts. Teardown deletes the owner policies before the org row,
// so a retry that begins after that point cannot resolve the owners any
// more; the record is what it recovers them from. The users themselves
// outlive the org, so ids are enough. Best-effort, like the notice itself.
func (d Service) recordNoticeRecipients(ctx context.Context, orgID string, owners []user.User) {
	ids := make([]string, 0, len(owners))
	for _, owner := range owners {
		ids = append(ids, owner.ID)
	}
	if err := audit.GetAuditor(ctx, orgID).LogWithAttrs(audit.OrgDeleteNoticeRecipientsEvent, audit.Target{
		ID:   orgID,
		Type: "organization",
	}, map[string]string{
		"owner_ids": strings.Join(ids, ","),
	}); err != nil {
		slog.WarnContext(ctx, "failed to record the delete notice recipients", "org_id", orgID, "error", err)
	}
}

// recoverRecipientsFromAudit loads the owner ids a failed earlier attempt
// recorded, for a retry that starts after the owner policies were already
// deleted. Best-effort: without a readable audit store there is nothing to
// recover and the notice goes unsent, which the send path logs.
func (d Service) recoverRecipientsFromAudit(ctx context.Context, orgID string) []user.User {
	logs, err := audit.GetService(ctx).List(ctx, audit.Filter{
		OrgID:  orgID,
		Action: string(audit.OrgDeleteNoticeRecipientsEvent),
	})
	if err != nil {
		slog.WarnContext(ctx, "failed to check audit records for the delete notice recipients", "org_id", orgID, "error", err)
		return nil
	}
	for _, l := range logs {
		raw := l.Metadata["owner_ids"]
		if raw == "" {
			continue
		}
		owners, err := d.userService.GetByIDs(ctx, strings.Split(raw, ","))
		if err != nil {
			slog.WarnContext(ctx, "failed to fetch the recorded delete notice recipients", "org_id", orgID, "error", err)
			return nil
		}
		return owners
	}
	return nil
}

// recoverForfeitFromAudit adds the forfeits a failed earlier teardown wrote
// audit records for, so the retry that completes the delete still reports
// them in the owner notice. It reconciles per billing account: only the
// newest record per account counts (a retried teardown can write the same
// forfeit twice), and an account that still holds a live balance is already
// counted by the collection pass, so its records are skipped. Best-effort:
// without a readable audit store the notice keeps only the live amounts.
func (d Service) recoverForfeitFromAudit(ctx context.Context, orgID string, notice *deleteNotice) {
	logs, err := audit.GetService(ctx).List(ctx, audit.Filter{
		OrgID:  orgID,
		Action: string(audit.BillingTokensForfeitedEvent),
	})
	if err != nil {
		slog.WarnContext(ctx, "failed to check audit records for forfeited tokens", "org_id", orgID, "error", err)
		return
	}
	// the list is newest first, so the first record per account wins
	seen := map[string]struct{}{}
	for _, l := range logs {
		accountID := l.Target.ID
		if accountID == "" {
			continue
		}
		if _, ok := seen[accountID]; ok {
			continue
		}
		seen[accountID] = struct{}{}
		if _, live := notice.Accounts[accountID]; live {
			continue
		}
		amount, _ := strconv.ParseInt(l.Metadata["amount"], 10, 64)
		purchased, _ := strconv.ParseInt(l.Metadata["purchased"], 10, 64)
		notice.Amount += amount
		notice.Purchased += purchased
	}
}

// resolveAccountTokens returns one account's balance and purchased share,
// from the caller's already-collected amounts when given, otherwise read
// fresh.
func (d Service) resolveAccountTokens(ctx context.Context, accountID string, amounts map[string]accountTokens) (accountTokens, error) {
	if amounts != nil {
		return amounts[accountID], nil
	}
	balance, err := d.creditService.GetBalance(ctx, accountID)
	if err != nil {
		return accountTokens{}, err
	}
	if balance <= 0 {
		return accountTokens{Balance: balance}, nil
	}
	bought, err := d.purchasedTokens(ctx, accountID, balance)
	if err != nil {
		return accountTokens{}, err
	}
	return accountTokens{Balance: balance, Purchased: bought}, nil
}

// purchasedTokens returns how many of the account's remaining tokens came
// from purchases. Complimentary tokens (plan starter grants and awards) are
// counted as spent first, so the purchased share is the smaller of the
// balance and everything ever bought.
func (d Service) purchasedTokens(ctx context.Context, accountID string, balance int64) (int64, error) {
	txns, err := d.creditService.List(ctx, credit.Filter{CustomerID: accountID})
	if err != nil {
		return 0, fmt.Errorf("failed to list token transactions of billing account[%s]: %w", accountID, err)
	}
	var bought int64
	for _, t := range txns {
		if t.Source != credit.SourceSystemBuyEvent {
			continue
		}
		switch t.Type {
		case credit.CreditType:
			bought += t.Amount
		case credit.DebitType:
			// a debit recorded against the buy source takes purchased
			// tokens back (a refund); it must not count as transferable
			bought -= t.Amount
		}
	}
	if bought < 0 {
		bought = 0
	}
	return min(bought, balance), nil
}

// sendDeleteNotices emails every org owner that the organization was
// deleted. The org is already gone at this point, so failures are logged
// and never returned.
func (d Service) sendDeleteNotices(ctx context.Context, org organization.Organization, notice deleteNotice) {
	if d.mailDialer == nil {
		slog.WarnContext(ctx, "no mail dialer configured, skipping the delete notices", "org_id", org.ID)
		return
	}
	if len(notice.Owners) == 0 {
		slog.WarnContext(ctx, "the organization was deleted but no owner could be notified", "org_id", org.ID, "amount", notice.Amount, "purchased", notice.Purchased)
		return
	}
	subjectTpl := d.deleteNoticeConfig.Subject
	if subjectTpl == "" {
		subjectTpl = defaultDeleteNoticeSubject
	}
	bodyTpl := d.deleteNoticeConfig.Body
	if bodyTpl == "" {
		bodyTpl = defaultDeleteNoticeBody
	}
	// the templates are the same for every owner; parse them once
	subjectTmpl, err := texttemplate.New("subject").Parse(subjectTpl)
	if err != nil {
		slog.WarnContext(ctx, "failed to parse the delete notice subject template", "org_id", org.ID, "error", err)
		return
	}
	bodyTmpl, err := htmltemplate.New("body").Parse(bodyTpl)
	if err != nil {
		slog.WarnContext(ctx, "failed to parse the delete notice body template", "org_id", org.ID, "error", err)
		return
	}

	deletedBy := deletedByFromContext(ctx)
	for _, owner := range notice.Owners {
		data := deleteNoticeData{
			Amount:    notice.Amount,
			Purchased: notice.Purchased,
			User:      owner,
			Org:       org,
			DeletedBy: deletedBy,
		}
		var subject, body bytes.Buffer
		if err := subjectTmpl.Execute(&subject, data); err != nil {
			slog.WarnContext(ctx, "failed to render the delete notice subject", "org_id", org.ID, "user_email", owner.Email, "error", err)
			continue
		}
		if err := bodyTmpl.Execute(&body, data); err != nil {
			slog.WarnContext(ctx, "failed to render the delete notice body", "org_id", org.ID, "user_email", owner.Email, "error", err)
			continue
		}

		msg := mail.NewMessage()
		msg.SetHeader("From", d.mailDialer.FromHeader())
		msg.SetHeader("To", owner.Email)
		msg.SetHeader("Subject", subject.String())
		msg.SetBody("text/html", body.String())
		if err := d.mailDialer.DialAndSend(msg); err != nil {
			slog.WarnContext(ctx, "failed to send the delete notice", "org_id", org.ID, "user_email", owner.Email, "error", err)
			continue
		}
		slog.InfoContext(ctx, "sent the organization delete notice", "org_id", org.ID, "user_email", owner.Email, "amount", notice.Amount)
	}
}

// deletedByFromContext names the caller who ran the delete, when the
// context carries one.
func deletedByFromContext(ctx context.Context) string {
	principal, ok := authenticate.GetPrincipalFromContext(ctx)
	if !ok || principal == nil {
		return ""
	}
	if principal.User != nil && principal.User.Email != "" {
		return principal.User.Email
	}
	if principal.ServiceUser != nil && principal.ServiceUser.Title != "" {
		return principal.ServiceUser.Title
	}
	return principal.ID
}
