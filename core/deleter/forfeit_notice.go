package deleter

import (
	"bytes"
	"context"
	"fmt"
	htmltemplate "html/template"
	"log/slog"
	"strconv"
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
	defaultForfeitNoticeSubject = "Unused tokens from your deleted organization"
	defaultForfeitNoticeBody    = `{{if .User.Title}}Hi {{.User.Title}},{{else}}Hi,{{end}}<br><br>Your organization <b>{{if .Org.Title}}{{.Org.Title}}{{else}}{{.Org.Name}}{{end}}</b> was deleted{{if .DeletedBy}} by <b>{{.DeletedBy}}</b>{{end}} with <b>{{.Amount}}</b> unused tokens remaining{{if and .Purchased (lt .Purchased .Amount)}}, of which <b>{{.Purchased}}</b> came from purchases{{end}}. {{if .Purchased}}Contact support to get the purchased amount transferred to your bank account.{{else}}These were complimentary tokens, so there is no amount to transfer.{{end}}`
)

type forfeitNoticeData struct {
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

// forfeitNotice is everything sendForfeitNotices needs once the org is gone.
// It has to be collected before teardown removes the owners and the token
// balances. Accounts keeps the per-account numbers so the teardown can audit
// them without reading the balances a second time.
type forfeitNotice struct {
	Amount    int64
	Purchased int64
	Accounts  map[string]accountTokens
	// Balances holds every account's balance, positive or not, so the
	// blocker check can reuse the reads
	Balances map[string]int64
	Owners   []user.User
}

// collectForfeitNotice sums the unused tokens the delete is about to forfeit
// and resolves the org owners to notify. It only reads; a failure here aborts
// the delete before anything is torn down.
//
// The amount is the whole remaining balance. Purchased is the share of it
// that came from purchases (source system.buy), with complimentary tokens
// (plan starter grants and awards) counted as spent first. Only the
// purchased share is transferable.
func (d Service) collectForfeitNotice(ctx context.Context, org organization.Organization, customers []customer.Customer) (forfeitNotice, error) {
	var total, purchased int64
	accounts := make(map[string]accountTokens, len(customers))
	balances := make(map[string]int64, len(customers))
	for _, c := range customers {
		balance, err := d.creditService.GetBalance(ctx, c.ID)
		if err != nil {
			return forfeitNotice{}, fmt.Errorf("failed to check token balance of billing account[%s]: %w", c.ID, err)
		}
		balances[c.ID] = balance
		if balance > 0 {
			bought, err := d.purchasedTokens(ctx, c.ID, balance)
			if err != nil {
				return forfeitNotice{}, err
			}
			total += balance
			purchased += bought
			accounts[c.ID] = accountTokens{Balance: balance, Purchased: bought}
		}
	}
	if total == 0 {
		return forfeitNotice{Accounts: accounts, Balances: balances}, nil
	}

	return forfeitNotice{
		Amount:    total,
		Purchased: purchased,
		Accounts:  accounts,
		Balances:  balances,
		Owners:    d.resolveOwners(ctx, org.ID),
	}, nil
}

// resolveOwners finds the users holding the org owner role. It is
// best-effort: the notice email must not make the delete depend on the
// policy machinery, so a failed lookup logs and returns no owners.
func (d Service) resolveOwners(ctx context.Context, orgID string) []user.User {
	ownerRole, err := d.roleService.Get(ctx, schema.RoleOrganizationOwner)
	if err != nil {
		slog.WarnContext(ctx, "failed to resolve the organization owner role for the forfeit notice", "org_id", orgID, "error", err)
		return nil
	}
	members, err := d.membershipService.ListPrincipalsByResource(ctx, orgID, schema.OrganizationNamespace, membership.MemberFilter{
		PrincipalType: schema.UserPrincipal,
		RoleIDs:       []string{ownerRole.ID},
	})
	if err != nil {
		slog.WarnContext(ctx, "failed to list the organization owners for the forfeit notice", "org_id", orgID, "error", err)
		return nil
	}
	ownerIDs := make([]string, 0, len(members))
	for _, m := range members {
		ownerIDs = append(ownerIDs, m.PrincipalID)
	}
	owners, err := d.userService.GetByIDs(ctx, ownerIDs)
	if err != nil {
		slog.WarnContext(ctx, "failed to fetch the organization owners for the forfeit notice", "org_id", orgID, "error", err)
		return nil
	}
	return owners
}

// recoverForfeitFromAudit adds the forfeits a failed earlier teardown wrote
// audit records for, so the retry that completes the delete still reports
// them in the owner notice. It reconciles per billing account: only the
// newest record per account counts (a retried teardown can write the same
// forfeit twice), and an account that still holds a live balance is already
// counted by the collection pass, so its records are skipped. Best-effort:
// without a readable audit store the notice keeps only the live amounts.
func (d Service) recoverForfeitFromAudit(ctx context.Context, orgID string, notice *forfeitNotice) {
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
	if notice.Amount > 0 && len(notice.Owners) == 0 {
		notice.Owners = d.resolveOwners(ctx, orgID)
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

// sendForfeitNotices emails every org owner that the delete forfeited unused
// tokens and that support can transfer the amount. The org is already gone at
// this point, so failures are logged and never returned.
func (d Service) sendForfeitNotices(ctx context.Context, org organization.Organization, notice forfeitNotice) {
	if d.mailDialer == nil {
		slog.WarnContext(ctx, "no mail dialer configured, skipping token forfeit notices", "org_id", org.ID)
		return
	}
	if len(notice.Owners) == 0 {
		slog.WarnContext(ctx, "tokens were forfeited but no owner could be notified", "org_id", org.ID, "amount", notice.Amount, "purchased", notice.Purchased)
		return
	}
	subjectTpl := d.forfeitNoticeConfig.Subject
	if subjectTpl == "" {
		subjectTpl = defaultForfeitNoticeSubject
	}
	bodyTpl := d.forfeitNoticeConfig.Body
	if bodyTpl == "" {
		bodyTpl = defaultForfeitNoticeBody
	}
	// the templates are the same for every owner; parse them once
	subjectTmpl, err := texttemplate.New("subject").Parse(subjectTpl)
	if err != nil {
		slog.WarnContext(ctx, "failed to parse token forfeit notice subject template", "org_id", org.ID, "error", err)
		return
	}
	bodyTmpl, err := htmltemplate.New("body").Parse(bodyTpl)
	if err != nil {
		slog.WarnContext(ctx, "failed to parse token forfeit notice body template", "org_id", org.ID, "error", err)
		return
	}

	deletedBy := deletedByFromContext(ctx)
	for _, owner := range notice.Owners {
		data := forfeitNoticeData{
			Amount:    notice.Amount,
			Purchased: notice.Purchased,
			User:      owner,
			Org:       org,
			DeletedBy: deletedBy,
		}
		var subject, body bytes.Buffer
		if err := subjectTmpl.Execute(&subject, data); err != nil {
			slog.WarnContext(ctx, "failed to render token forfeit notice subject", "org_id", org.ID, "user_email", owner.Email, "error", err)
			continue
		}
		if err := bodyTmpl.Execute(&body, data); err != nil {
			slog.WarnContext(ctx, "failed to render token forfeit notice body", "org_id", org.ID, "user_email", owner.Email, "error", err)
			continue
		}

		msg := mail.NewMessage()
		msg.SetHeader("From", d.mailDialer.FromHeader())
		msg.SetHeader("To", owner.Email)
		msg.SetHeader("Subject", subject.String())
		msg.SetBody("text/html", body.String())
		if err := d.mailDialer.DialAndSend(msg); err != nil {
			slog.WarnContext(ctx, "failed to send token forfeit notice", "org_id", org.ID, "user_email", owner.Email, "error", err)
			continue
		}
		slog.InfoContext(ctx, "sent token forfeit notice", "org_id", org.ID, "user_email", owner.Email, "amount", notice.Amount)
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
