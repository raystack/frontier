package e2e_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/suite"

	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/authenticate/strategy"
	testusers "github.com/raystack/frontier/core/authenticate/test_users"
	"github.com/raystack/frontier/core/consent"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/logger"
	"github.com/raystack/frontier/pkg/server"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/frontier/test/e2e/testbench"
)

// The cross-cutting matrix for RFC 0002 (docs/rfcs/0002-explicit-consent-at-signup.md):
// the flow intent, the consent rules and every strategy at once, through a real
// server. Per-rule unit coverage lives with the code it covers; this file is the
// coverage none of those own, because it spans all three axes.
//
// Two suites, because app.consent is read at boot: one with the feature on, one
// with it off. What is deliberately absent is recorded on
// TestPassKeyFinishIsNotReachableEndToEnd below.

const (
	// mail OTP runs through test users, not SMTP: any address on this domain gets
	// testbench.TestOTP, so a signup is two RPCs with no mailbox in between.
	consentTestUserDomain = "raystack.org"

	// mockoidc is registered under this name, which is also the flow's Method and
	// so what lands in auth_strategy.
	consentOIDCStrategy = "mock"

	// Authenticate is skip-listed, so the handler reads the IP off this header
	// itself rather than off the context.
	consentClientIPHeader = "X-Forwarded-For"
	consentClientIP       = "203.0.113.7"
)

// The documents the enabled suite configures. Two of them, so "the complete
// set" and "an incomplete set" are different requests rather than the same one
// with and without an id.
const (
	consentTermsID      = "terms_of_service"
	consentTermsTitle   = "Terms & Conditions"
	consentTermsVersion = "2026-04-01"
	consentTermsURL     = "https://example.org/legal/terms/2026-04-01"

	consentPrivacyID      = "privacy_policy"
	consentPrivacyTitle   = "Privacy Policy"
	consentPrivacyVersion = "2026-04-01"
	consentPrivacyURL     = "https://example.org/legal/privacy/2026-04-01"
)

// completeConsentIDs covers every configured document, which is what a signup
// needs. Deliberately not sorted: the service deduplicates and sorts, and a
// client sending its own order must still be accepted.
func completeConsentIDs() []string {
	return []string{consentTermsID, consentPrivacyID}
}

func enabledConsentConfig() consent.Config {
	return consent.Config{
		Enabled: true,
		Documents: map[string]consent.DocumentConfig{
			consentTermsID: {
				Title:   consentTermsTitle,
				Version: consentTermsVersion,
				URL:     consentTermsURL,
			},
			consentPrivacyID: {
				Title:   consentPrivacyTitle,
				Version: consentPrivacyVersion,
				URL:     consentPrivacyURL,
			},
		},
	}
}

// consentHarness is the machinery both suites share: a frontier server with every
// flow-based strategy wired up, the mocks an OIDC round trip needs, and a direct
// database handle for the two things the API deliberately does not expose —
// whether a user row exists, and what a consent record holds.
type consentHarness struct {
	suite.Suite
	testBench      *testbench.TestBench
	dbc            *db.Client
	mockOIDCServer *mockoidc.MockOIDC
	callbackServer *http.Server
	// allowlisted in config, so it survives SanitizeReturnToURL and becomes the
	// flow's finish URL.
	returnToURL string
	// a fresh code per round trip: mockoidc will not exchange one twice.
	oidcCodes int
}

func (s *consentHarness) start(consentConfig consent.Config) {
	connectPort, err := testbench.GetFreePort()
	s.Require().NoError(err)
	callbackPort, err := testbench.GetFreePort()
	s.Require().NoError(err)

	s.mockOIDCServer, err = mockoidc.Run()
	s.Require().NoError(err)

	callbackURL := "http://localhost:" + strconv.Itoa(callbackPort) + "/callback"
	s.returnToURL = "http://localhost:" + strconv.Itoa(callbackPort) + "/return"

	// where the provider redirects at the end of the authorization step. Nothing
	// asserts on it, but the redirect has to land somewhere that answers.
	mux := http.NewServeMux()
	mux.Handle("/callback", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	s.callbackServer = &http.Server{
		Addr:              "localhost:" + strconv.Itoa(callbackPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := s.callbackServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.T().Log("callback server stopped:", err)
		}
	}()

	appConfig := &config.Frontier{
		Log: logger.Config{
			Level: "error",
		},
		App: server.Config{
			Host:    "localhost",
			Connect: server.ConnectConfig{Port: connectPort},
			Authentication: authenticate.Config{
				Session: authenticate.SessionConfig{
					HashSecretKey:  "hash-secret-should-be-32-chars--",
					BlockSecretKey: "hash-secret-should-be-32-chars--",
					Validity:       time.Hour,
					// The struct is built here rather than loaded, so the
					// default tags do not apply and the header has to be named.
					Headers: authenticate.SessionMetadataHeaders{
						ClientIP: consentClientIPHeader,
					},
				},
				OIDCConfig: map[string]authenticate.OIDCConfig{
					consentOIDCStrategy: {
						ClientID:     s.mockOIDCServer.Config().ClientID,
						ClientSecret: s.mockOIDCServer.Config().ClientSecret,
						IssuerUrl:    s.mockOIDCServer.Issuer(),
					},
				},
				CallbackURLs:           []string{callbackURL},
				AuthorizedRedirectURLs: []string{s.returnToURL},
				MailOTP: authenticate.MailOTPConfig{
					Subject:  "{{.Otp}}",
					Body:     "{{.Otp}}",
					Validity: 10 * time.Minute,
				},
				TestUsers: testusers.Config{
					Enabled: true,
					Domain:  consentTestUserDomain,
					OTP:     testbench.TestOTP,
				},
				// Passkey is off unless the relying party is configured, and
				// StartFlow rejects a strategy it does not support, so the
				// passkey half of the matrix needs this block to exist.
				PassKey: authenticate.PassKeyConfig{
					RPDisplayName: "Frontier E2E",
					RPID:          "localhost",
					RPOrigins:     []string{"http://localhost"},
				},
			},
			Consent: consentConfig,
		},
	}

	s.testBench, err = testbench.Init(appConfig)
	s.Require().NoError(err)

	// testbench.Init fills this in, so the suite can read the rows the API does
	// not serve.
	s.dbc, err = db.New(appConfig.DB)
	s.Require().NoError(err)
}

func (s *consentHarness) stop() {
	s.Require().NoError(s.dbc.Close())
	s.Require().NoError(s.callbackServer.Shutdown(context.Background()))
	// Close SIGINTs the test process to stop the server, so it goes last.
	s.Require().NoError(s.testBench.Close())
	s.Require().NoError(s.mockOIDCServer.Shutdown())
}

// clientContext carries the forwarded-for header on every call, so an accepted
// signup records the address it came from.
func (s *consentHarness) clientContext() context.Context {
	return testbench.ContextWithHeaders(context.Background(), map[string]string{
		consentClientIPHeader: consentClientIP,
	})
}

func (s *consentHarness) authenticate(ctx context.Context, req *frontierv1beta1.AuthenticateRequest) (*connect.Response[frontierv1beta1.AuthenticateResponse], error) {
	return s.testBench.Client.Authenticate(ctx, connect.NewRequest(req))
}

// startMailOTP asks for a one time password. Test users are configured, so the
// code is always testbench.TestOTP and no SMTP server is involved.
func (s *consentHarness) startMailOTP(ctx context.Context, email string, intent frontierv1beta1.FlowIntent, acceptedIDs []string) (string, error) {
	resp, err := s.authenticate(ctx, &frontierv1beta1.AuthenticateRequest{
		StrategyName:        strategy.MailOTPAuthMethod,
		Email:               email,
		FlowIntent:          intent,
		AcceptedDocumentIds: acceptedIDs,
	})
	if err != nil {
		return "", err
	}
	return resp.Msg.GetState(), nil
}

func (s *consentHarness) finishMailOTP(ctx context.Context, state string) (*connect.Response[frontierv1beta1.AuthCallbackResponse], error) {
	return s.testBench.Client.AuthCallback(ctx, connect.NewRequest(&frontierv1beta1.AuthCallbackRequest{
		StrategyName: strategy.MailOTPAuthMethod,
		Code:         testbench.TestOTP,
		State:        state,
	}))
}

// signUpWithMailOTP creates an account the way a client would, so cases needing
// an existing address test one the feature itself produced.
func (s *consentHarness) signUpWithMailOTP(ctx context.Context, email string, acceptedIDs []string) {
	state, err := s.startMailOTP(ctx, email, frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP, acceptedIDs)
	s.Require().NoError(err)
	_, err = s.finishMailOTP(ctx, state)
	s.Require().NoError(err)
}

// startOIDC runs Authenticate for the mock provider and then the browser half of
// the flow, handing back the state and code AuthCallback needs. returnTo becomes
// the finish URL, which only a successful callback uses.
func (s *consentHarness) startOIDC(ctx context.Context, email, returnTo string, intent frontierv1beta1.FlowIntent, acceptedIDs []string) (string, string, error) {
	authResp, err := s.authenticate(ctx, &frontierv1beta1.AuthenticateRequest{
		StrategyName:        consentOIDCStrategy,
		Email:               email,
		ReturnTo:            returnTo,
		FlowIntent:          intent,
		AcceptedDocumentIds: acceptedIDs,
	})
	if err != nil {
		return "", "", err
	}

	endpoint := authResp.Msg.GetEndpoint()
	s.Require().NotEmpty(endpoint)
	parsedEndpoint, err := url.Parse(endpoint)
	s.Require().NoError(err)

	// the account is created for the address the provider asserts, not the one the
	// client asked with, so queue a profile per case.
	s.oidcCodes++
	code := "e2e-code-" + strconv.Itoa(s.oidcCodes)
	s.mockOIDCServer.QueueUser(&mockoidc.MockUser{
		Subject:           email,
		Email:             email,
		EmailVerified:     true,
		PreferredUsername: email,
	})
	s.mockOIDCServer.QueueCode(code)

	res, err := http.Get(endpoint) //nolint:gosec // the endpoint comes from the server under test
	s.Require().NoError(err)
	defer res.Body.Close() //nolint:errcheck
	s.Require().Equal(http.StatusOK, res.StatusCode)

	return parsedEndpoint.Query().Get("state"), code, nil
}

func (s *consentHarness) finishOIDC(ctx context.Context, state, code string) (*connect.Response[frontierv1beta1.AuthCallbackResponse], error) {
	return s.testBench.Client.AuthCallback(ctx, connect.NewRequest(&frontierv1beta1.AuthCallbackRequest{
		Code:  code,
		State: state,
	}))
}

// userCount reports how many user rows exist for an address. Zero is the
// assertion that matters: it is what "nothing was written" means for a
// rejected signup, and no endpoint reports it.
func (s *consentHarness) userCount(email string) int {
	var count int
	s.Require().NoError(s.dbc.GetContext(context.Background(), &count,
		"SELECT count(*) FROM users WHERE email = $1", email))
	return count
}

// consentsFor reads the records written for an address. The repository has a
// Create and nothing else by design, so this is the only way to see them, and
// it is how reporting reads them too.
func (s *consentHarness) consentsFor(email string) []postgres.UserConsent {
	var rows []postgres.UserConsent
	s.Require().NoError(s.dbc.SelectContext(context.Background(), &rows,
		"SELECT * FROM user_consents WHERE user_email = $1 ORDER BY created_at", email))
	return rows
}

func (s *consentHarness) documentsIn(row postgres.UserConsent) []postgres.ConsentDocument {
	var documents []postgres.ConsentDocument
	s.Require().NoError(json.Unmarshal(row.Documents, &documents))
	return documents
}

// assertSessionStarted checks that the callback really logged somebody in,
// rather than returning a response that happens to carry no error.
func (s *consentHarness) assertSessionStarted(resp *connect.Response[frontierv1beta1.AuthCallbackResponse], email string) {
	setCookie := resp.Header().Get("Set-Cookie")
	s.Require().NotEmpty(setCookie)
	cookie := strings.SplitN(setCookie, ";", 2)[0]

	getUserResp, err := s.testBench.Client.GetCurrentUser(
		testbench.ContextWithAuth(context.Background(), cookie),
		connect.NewRequest(&frontierv1beta1.GetCurrentUserRequest{}))
	s.Require().NoError(err)
	s.Assert().Equal(email, getUserResp.Msg.GetUser().GetEmail())
}

// assertSignupRecorded checks the whole result of an accepted signup: the
// account exists, exactly one consent record covers it, and the record holds
// the config snapshot of every document plus the act that accepted them.
func (s *consentHarness) assertSignupRecorded(email, authStrategy string, before time.Time) {
	s.Assert().Equal(1, s.userCount(email))

	rows := s.consentsFor(email)
	s.Require().Len(rows, 1)
	row := rows[0]

	s.Assert().NotEmpty(row.UserID)
	s.Assert().Equal(email, row.UserEmail)
	s.Assert().Equal(consent.SourceSignup, row.Source)
	s.Assert().Equal(authStrategy, row.AuthStrategy.String)
	s.Assert().Equal(consentClientIP, row.IPAddress.String)

	// consented_at is the acceptance at flow start, not the write, so it sits
	// between the start of the case and now.
	s.Assert().False(row.ConsentedAt.Before(before), "consented_at should not predate the flow")
	s.Assert().False(row.ConsentedAt.After(row.CreatedAt), "consented_at should not postdate the write")

	// the client sent ids and nothing else, so everything else here came from
	// config. Ordered by id, which is why privacy_policy comes first.
	s.Assert().Equal([]postgres.ConsentDocument{
		{ID: consentPrivacyID, Title: consentPrivacyTitle, Version: consentPrivacyVersion, URL: consentPrivacyURL},
		{ID: consentTermsID, Title: consentTermsTitle, Version: consentTermsVersion, URL: consentTermsURL},
	}, s.documentsIn(row))
}

// assertNothingWritten is the other half of every rejection: a rejected signup
// leaves no account and no record, whichever gate rejected it.
func (s *consentHarness) assertNothingWritten(email string) {
	s.Assert().Equal(0, s.userCount(email), "a rejected signup must not create a user")
	s.Assert().Empty(s.consentsFor(email), "a rejected signup must not write a consent record")
}

// ---------------------------------------------------------------------------
// app.consent enabled
// ---------------------------------------------------------------------------

type ConsentRegressionTestSuite struct {
	consentHarness
}

func (s *ConsentRegressionTestSuite) SetupSuite() {
	s.start(enabledConsentConfig())
}

func (s *ConsentRegressionTestSuite) TearDownSuite() {
	s.stop()
}

func (s *ConsentRegressionTestSuite) TestListConsentDocumentsWithTheFeatureEnabled() {
	// Unauthenticated on purpose: the ids are an input to an unauthenticated
	// Authenticate, so requiring a session to learn what to accept before the
	// account exists would be a cycle. A bare context proves it.
	resp, err := s.testBench.Client.ListConsentDocuments(context.Background(),
		connect.NewRequest(&frontierv1beta1.ListConsentDocumentsRequest{}))
	s.Require().NoError(err)

	documents := resp.Msg.GetDocuments()
	s.Require().Len(documents, 2)

	// Ordered by id, so the response is stable across boots even though the
	// config behind it is a map.
	s.Assert().Equal(consentPrivacyID, documents[0].GetId())
	s.Assert().Equal(consentPrivacyTitle, documents[0].GetTitle())
	s.Assert().Equal(consentPrivacyVersion, documents[0].GetVersion())
	s.Assert().Equal(consentPrivacyURL, documents[0].GetUrl())

	s.Assert().Equal(consentTermsID, documents[1].GetId())
	s.Assert().Equal(consentTermsTitle, documents[1].GetTitle())
	s.Assert().Equal(consentTermsVersion, documents[1].GetVersion())
	s.Assert().Equal(consentTermsURL, documents[1].GetUrl())
}

// TestMailOTPIntentAndConsent walks the three intents against an address that does
// and does not have an account, at both enforcement points. Mail OTP knows the
// address up front, so most of these are rejected before a code is issued.
func (s *ConsentRegressionTestSuite) TestMailOTPIntentAndConsent() {
	ctx := s.clientContext()
	existing := "consent-mailotp-existing@" + consentTestUserDomain

	s.Run("1. a signup with the complete set creates the account and records the consent", func() {
		before := time.Now().UTC()

		state, err := s.startMailOTP(ctx, existing, frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP, completeConsentIDs())
		s.Require().NoError(err)

		callbackResp, err := s.finishMailOTP(ctx, state)
		s.Require().NoError(err)
		s.assertSessionStarted(callbackResp, existing)

		s.assertSignupRecorded(existing, strategy.MailOTPAuthMethod, before)
	})

	s.Run("2. a signup for an address that already has an account is rejected at flow start", func() {
		_, err := s.startMailOTP(ctx, existing, frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP, completeConsentIDs())
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeAlreadyExists, connect.CodeOf(err))

		// the account and its record are untouched: no second row, no rewrite
		s.Assert().Len(s.consentsFor(existing), 1)
	})

	s.Run("3. a signup with an incomplete set is rejected before an OTP is issued", func() {
		email := "consent-mailotp-incomplete@" + consentTestUserDomain

		_, err := s.startMailOTP(ctx, email, frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP, []string{consentTermsID})
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))
		// the wrapped error names the missing ids; the response must not
		s.Assert().NotContains(err.Error(), consentPrivacyID)

		s.assertNothingWritten(email)
	})

	s.Run("4. a signup naming a document this deployment does not configure is rejected", func() {
		email := "consent-mailotp-unknown@" + consentTestUserDomain

		// complete plus one unknown id, so the unknown-id rule fires first
		_, err := s.startMailOTP(ctx, email, frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP,
			append(completeConsentIDs(), "cookie_policy"))
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))

		s.assertNothingWritten(email)
	})

	s.Run("5. a login for an address with no account is rejected at flow start", func() {
		email := "consent-mailotp-nologin@" + consentTestUserDomain

		_, err := s.startMailOTP(ctx, email, frontierv1beta1.FlowIntent_FLOW_INTENT_LOGIN, nil)
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeNotFound, connect.CodeOf(err))

		// the point of the gate: a login no longer creates the account
		s.assertNothingWritten(email)
	})

	s.Run("6. a login for an existing account signs in and writes no second record", func() {
		state, err := s.startMailOTP(ctx, existing, frontierv1beta1.FlowIntent_FLOW_INTENT_LOGIN, nil)
		s.Require().NoError(err)

		callbackResp, err := s.finishMailOTP(ctx, state)
		s.Require().NoError(err)
		s.assertSessionStarted(callbackResp, existing)

		s.Assert().Len(s.consentsFor(existing), 1)
	})

	s.Run("7. accepted ids sent with a login intent are rejected as a client bug", func() {
		_, err := s.startMailOTP(ctx, existing, frontierv1beta1.FlowIntent_FLOW_INTENT_LOGIN, completeConsentIDs())
		s.Require().Error(err)
		// InvalidArgument, not FailedPrecondition: the consent is fine, the
		// request is malformed
		s.Assert().Equal(connect.CodeInvalidArgument, connect.CodeOf(err))
	})

	s.Run("8. an unspecified intent with the complete set creates the account and records the consent", func() {
		email := "consent-mailotp-unspecified@" + consentTestUserDomain
		before := time.Now().UTC()

		// without an intent the check splits: Resolve at flow start, ResolveAll
		// again at user creation, the first point frontier knows the user is new
		state, err := s.startMailOTP(ctx, email, frontierv1beta1.FlowIntent_FLOW_INTENT_UNSPECIFIED, completeConsentIDs())
		s.Require().NoError(err)

		callbackResp, err := s.finishMailOTP(ctx, state)
		s.Require().NoError(err)
		s.assertSessionStarted(callbackResp, email)

		s.assertSignupRecorded(email, strategy.MailOTPAuthMethod, before)
	})

	s.Run("9. an unspecified intent with no consent passes flow start and is rejected at user creation", func() {
		email := "consent-mailotp-noconsent@" + consentTestUserDomain

		// the second enforcement point alone: no completeness rule applies at
		// flow start without an intent, so the OTP is issued...
		state, err := s.startMailOTP(ctx, email, frontierv1beta1.FlowIntent_FLOW_INTENT_UNSPECIFIED, nil)
		s.Require().NoError(err)

		// ...and the rejection lands where the account would have been created.
		_, err = s.finishMailOTP(ctx, state)
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))

		s.assertNothingWritten(email)
	})

	s.Run("10. an unspecified intent signs an existing user in whatever the flow carries", func() {
		// consent is a rule about creating accounts, so an existing user is
		// never blocked and never gets a record
		state, err := s.startMailOTP(ctx, existing, frontierv1beta1.FlowIntent_FLOW_INTENT_UNSPECIFIED, nil)
		s.Require().NoError(err)

		callbackResp, err := s.finishMailOTP(ctx, state)
		s.Require().NoError(err)
		s.assertSessionStarted(callbackResp, existing)

		s.Assert().Len(s.consentsFor(existing), 1)
	})
}

// TestOIDCIntentAndConsent is the same matrix through applyOIDC, and the reason
// the gates exist in two places: the address is unknown until the provider
// asserts it, so nothing can be checked at flow start.
func (s *ConsentRegressionTestSuite) TestOIDCIntentAndConsent() {
	ctx := s.clientContext()
	existing := "consent-oidc-existing@example.org"

	s.Run("1. a signup with the complete set creates the account and records the consent", func() {
		before := time.Now().UTC()

		state, code, err := s.startOIDC(ctx, existing, s.returnToURL,
			frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP, completeConsentIDs())
		s.Require().NoError(err)

		callbackResp, err := s.finishOIDC(ctx, state, code)
		s.Require().NoError(err)
		s.assertSessionStarted(callbackResp, existing)

		// auth_strategy is the flow's own method, which for OIDC is the provider
		s.assertSignupRecorded(existing, consentOIDCStrategy, before)
	})

	s.Run("2. a signup for an address that already has an account is rejected at the callback", func() {
		// Flow start cannot reject this one: the provider has not spoken yet.
		state, code, err := s.startOIDC(ctx, existing, s.returnToURL,
			frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP, completeConsentIDs())
		s.Require().NoError(err)

		// the rejection is the answer: the application's callback page is what
		// calls this RPC, and decides where the user goes next
		callbackResp, err := s.finishOIDC(ctx, state, code)
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeAlreadyExists, connect.CodeOf(err))
		s.Assert().Nil(callbackResp, "a rejected callback must not start a session")

		s.Assert().Len(s.consentsFor(existing), 1)
	})

	s.Run("3. a login for an address with no account is rejected at the callback", func() {
		email := "consent-oidc-nologin@example.org"

		state, code, err := s.startOIDC(ctx, email, s.returnToURL,
			frontierv1beta1.FlowIntent_FLOW_INTENT_LOGIN, nil)
		s.Require().NoError(err)

		_, err = s.finishOIDC(ctx, state, code)
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeNotFound, connect.CodeOf(err))

		s.assertNothingWritten(email)
	})

	s.Run("4. an unspecified intent with no consent is rejected at the callback", func() {
		email := "consent-oidc-noconsent@example.org"

		state, code, err := s.startOIDC(ctx, email, s.returnToURL,
			frontierv1beta1.FlowIntent_FLOW_INTENT_UNSPECIFIED, nil)
		s.Require().NoError(err)

		_, err = s.finishOIDC(ctx, state, code)
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))

		// the bare sentinel goes out: the client owns the copy, and the wrapped
		// error naming the missing documents stays in the log
		s.Assert().NotContains(connect.CodeOf(err).String(), consentPrivacyID)
		s.Assert().NotContains(err.Error(), consentPrivacyID)

		s.assertNothingWritten(email)
	})

	s.Run("5. a return_to on the flow changes nothing about the rejection", func() {
		email := "consent-oidc-noreturnto@example.org"

		// only a successful callback uses the finish URL, so an allowlisted
		// return_to and no return_to answer a rejection identically
		state, code, err := s.startOIDC(ctx, email, "",
			frontierv1beta1.FlowIntent_FLOW_INTENT_LOGIN, nil)
		s.Require().NoError(err)

		_, err = s.finishOIDC(ctx, state, code)
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeNotFound, connect.CodeOf(err))

		s.assertNothingWritten(email)
	})
}

// TestPassKeyIntentAndConsent covers the reachable half: the gates at flow start,
// where the intent replaced the register-or-login guess. The ceremony itself is
// out of reach — see TestPassKeyFinishIsNotReachableEndToEnd.
func (s *ConsentRegressionTestSuite) TestPassKeyIntentAndConsent() {
	ctx := s.clientContext()

	// an account for the cases that need one; mail OTP is the cheapest way to
	// produce a real one, and this suite has already covered that path
	existing := "consent-passkey-existing@" + consentTestUserDomain
	s.signUpWithMailOTP(ctx, existing, completeConsentIDs())

	s.Run("1. a signup with an incomplete set is rejected before the ceremony starts", func() {
		email := "consent-passkey-incomplete@" + consentTestUserDomain

		// the consent gate runs before the passkey branch, so no lookup, no challenge
		_, err := s.authenticate(ctx, &frontierv1beta1.AuthenticateRequest{
			StrategyName:        strategy.PasskeyAuthMethod,
			Email:               email,
			FlowIntent:          frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP,
			AcceptedDocumentIds: []string{consentPrivacyID},
		})
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))

		s.assertNothingWritten(email)
	})

	s.Run("2. a signup for an address that already has an account is rejected at flow start", func() {
		_, err := s.authenticate(ctx, &frontierv1beta1.AuthenticateRequest{
			StrategyName:        strategy.PasskeyAuthMethod,
			Email:               existing,
			FlowIntent:          frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP,
			AcceptedDocumentIds: completeConsentIDs(),
		})
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeAlreadyExists, connect.CodeOf(err))
	})

	s.Run("3. a login for an address with no account is rejected at flow start", func() {
		email := "consent-passkey-nologin@" + consentTestUserDomain

		// the case the intent was added for: the ceremony used to be guessed, and
		// a passkey login would then create the account
		_, err := s.authenticate(ctx, &frontierv1beta1.AuthenticateRequest{
			StrategyName: strategy.PasskeyAuthMethod,
			Email:        email,
			FlowIntent:   frontierv1beta1.FlowIntent_FLOW_INTENT_LOGIN,
		})
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeNotFound, connect.CodeOf(err))

		s.assertNothingWritten(email)
	})

	s.Run("4. a login for an account with no stored passkey fails rather than registering one", func() {
		// the login gate passes and the intent picks the login ceremony, which has
		// nothing to challenge against; it must fail rather than fall back to
		// registering, as the guess used to
		_, err := s.authenticate(ctx, &frontierv1beta1.AuthenticateRequest{
			StrategyName: strategy.PasskeyAuthMethod,
			Email:        existing,
			FlowIntent:   frontierv1beta1.FlowIntent_FLOW_INTENT_LOGIN,
		})
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))
		// case 1 is a FailedPrecondition too, so the code alone would not tell
		// this apart from the consent gate firing first; the message is what does
		s.Assert().Contains(err.Error(), authenticate.ErrInvalidMethod.Error())
		s.Assert().Len(s.consentsFor(existing), 1)
	})

	s.Run("5. a signup that passes both gates starts the registration ceremony and creates nothing yet", func() {
		email := "consent-passkey-signup@" + consentTestUserDomain

		resp, err := s.authenticate(ctx, &frontierv1beta1.AuthenticateRequest{
			StrategyName:        strategy.PasskeyAuthMethod,
			Email:               email,
			FlowIntent:          frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP,
			AcceptedDocumentIds: completeConsentIDs(),
		})
		s.Require().NoError(err)
		s.Assert().NotEmpty(resp.Msg.GetState())

		// The intent, not a lookup, is what picked the ceremony.
		s.Require().NotNil(resp.Msg.GetStateOptions())
		s.Assert().Equal(strategy.PasskeyRegisterType,
			resp.Msg.GetStateOptions().GetFields()["type"].GetStringValue())

		// the account is created when the ceremony finishes, so nothing exists yet
		// and the carried consent is not written early
		s.assertNothingWritten(email)
	})
}

// TestPassKeyFinishIsNotReachableEndToEnd records the one hole in the matrix, so
// it shows up in a test run rather than only in review.
//
// Both finish methods reach getOrCreateUser, and a first-time passkey login is a
// signup, so both belong here. The register method validates the attestation
// before the gate, so the gate sits behind a ceremony this suite cannot perform;
// mocking the signature would test the mock. The login method gates first, but
// reaching it needs an account carrying a passkey_credentials blob, which only
// that same ceremony writes.
//
// Both gates are the same getOrCreateUser call the covered strategies use, so the
// risk is low — but nothing, here or at unit level, drives either method. Closing
// it needs a virtual authenticator, which is a new dependency of its own.
func (s *ConsentRegressionTestSuite) TestPassKeyFinishIsNotReachableEndToEnd() {
	s.T().Skip("passkey finish needs a WebAuthn authenticator; see the comment above " +
		"for why neither method is reachable and what closing the gap would take")
}

// TestConsentInsertFailureRollsBackTheUser is the invariant the transaction exists
// for: a user row without a consent record is impossible. The failure is injected
// with a trigger, because nothing in the API can produce one — a bad payload is
// rejected before the transaction opens, and a well-formed write satisfies every
// constraint on the table.
func (s *ConsentRegressionTestSuite) TestConsentInsertFailureRollsBackTheUser() {
	ctx := s.clientContext()
	// The address is spelled out in the trigger below as well; Postgres has no
	// parameters in a function body, and a formatted query would trip the SQL
	// safety linters for no benefit in a fixture.
	email := "consent-rollback@raystack.org"

	_, err := s.dbc.ExecContext(ctx, `
CREATE OR REPLACE FUNCTION e2e_reject_user_consent_insert()
  RETURNS TRIGGER AS $$
BEGIN
    IF NEW.user_email = 'consent-rollback@raystack.org' THEN
        -- no ERRCODE, so this is a plain insert failure rather than one the
        -- repository has a named error for
        RAISE EXCEPTION 'e2e injected consent insert failure';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_e2e_reject_user_consent_insert
    BEFORE INSERT ON user_consents
    FOR EACH ROW EXECUTE FUNCTION e2e_reject_user_consent_insert();`)
	s.Require().NoError(err)

	dropped := false
	dropTrigger := func() {
		if dropped {
			return
		}
		_, err := s.dbc.ExecContext(context.Background(), `
DROP TRIGGER IF EXISTS trg_e2e_reject_user_consent_insert ON user_consents;
DROP FUNCTION IF EXISTS e2e_reject_user_consent_insert();`)
		s.Require().NoError(err)
		dropped = true
	}
	defer dropTrigger()

	s.Run("1. the signup fails and leaves neither row behind", func() {
		state, err := s.startMailOTP(ctx, email, frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP, completeConsentIDs())
		s.Require().NoError(err)

		_, err = s.finishMailOTP(ctx, state)
		s.Require().Error(err)

		// the user insert had already run when the consent insert failed; both gone
		s.assertNothingWritten(email)
	})

	s.Run("2. and the same signup succeeds once the failure is removed", func() {
		// proof the trigger was the only obstacle: the rollback left nothing
		// behind for a retry to collide with
		dropTrigger()
		before := time.Now().UTC()

		state, err := s.startMailOTP(ctx, email, frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP, completeConsentIDs())
		s.Require().NoError(err)

		callbackResp, err := s.finishMailOTP(ctx, state)
		s.Require().NoError(err)
		s.assertSessionStarted(callbackResp, email)

		s.assertSignupRecorded(email, strategy.MailOTPAuthMethod, before)
	})
}

func TestEndToEndConsentRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(ConsentRegressionTestSuite))
}

// ---------------------------------------------------------------------------
// app.consent disabled
// ---------------------------------------------------------------------------

// ConsentDisabledRegressionTestSuite is the promise the RFC makes to deployments
// that never turn the feature on: nothing changes. It needs its own server
// because app.consent is read at boot. The intent is a separate axis and still
// applies, which is the point of the last case.
type ConsentDisabledRegressionTestSuite struct {
	consentHarness
}

func (s *ConsentDisabledRegressionTestSuite) SetupSuite() {
	s.start(consent.Config{})
}

func (s *ConsentDisabledRegressionTestSuite) TearDownSuite() {
	s.stop()
}

func (s *ConsentDisabledRegressionTestSuite) TestListConsentDocumentsWithTheFeatureDisabled() {
	resp, err := s.testBench.Client.ListConsentDocuments(context.Background(),
		connect.NewRequest(&frontierv1beta1.ListConsentDocumentsRequest{}))
	// An empty list, not an error, so one client build works against both kinds
	// of deployment: no documents means no checkbox.
	s.Require().NoError(err)
	s.Assert().Empty(resp.Msg.GetDocuments())
}

func (s *ConsentDisabledRegressionTestSuite) TestSignupBehavesAsItDidBeforeConsentExisted() {
	ctx := s.clientContext()

	s.Run("1. an unspecified intent with no ids creates the account, as it always did", func() {
		email := "noconsent-unspecified@" + consentTestUserDomain

		state, err := s.startMailOTP(ctx, email, frontierv1beta1.FlowIntent_FLOW_INTENT_UNSPECIFIED, nil)
		s.Require().NoError(err)
		callbackResp, err := s.finishMailOTP(ctx, state)
		s.Require().NoError(err)
		s.assertSessionStarted(callbackResp, email)

		s.Assert().Equal(1, s.userCount(email))
		s.Assert().Empty(s.consentsFor(email), "a disabled deployment records no consent")
	})

	s.Run("2. an unspecified intent signs the same address in rather than creating a second account", func() {
		email := "noconsent-unspecified@" + consentTestUserDomain

		state, err := s.startMailOTP(ctx, email, frontierv1beta1.FlowIntent_FLOW_INTENT_UNSPECIFIED, nil)
		s.Require().NoError(err)
		callbackResp, err := s.finishMailOTP(ctx, state)
		s.Require().NoError(err)
		s.assertSessionStarted(callbackResp, email)

		s.Assert().Equal(1, s.userCount(email))
	})

	s.Run("3. accepted ids are ignored rather than rejected", func() {
		email := "noconsent-withids@" + consentTestUserDomain

		// a client built for a consent deployment still works against one that is
		// not: unknown ids are neither rejected nor recorded
		state, err := s.startMailOTP(ctx, email, frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP,
			[]string{consentTermsID, consentPrivacyID})
		s.Require().NoError(err)
		callbackResp, err := s.finishMailOTP(ctx, state)
		s.Require().NoError(err)
		s.assertSessionStarted(callbackResp, email)

		s.Assert().Equal(1, s.userCount(email))
		s.Assert().Empty(s.consentsFor(email))
	})

	s.Run("4. accepted ids sent with a login intent are ignored too", func() {
		// enabled, this is the one shape the handler turns down on its own;
		// disabled it writes no record under any intent, so there is nothing an
		// id could refer to and the same client build has to keep working
		existing := "noconsent-unspecified@" + consentTestUserDomain

		state, err := s.startMailOTP(ctx, existing, frontierv1beta1.FlowIntent_FLOW_INTENT_LOGIN,
			[]string{consentTermsID, consentPrivacyID})
		s.Require().NoError(err)
		callbackResp, err := s.finishMailOTP(ctx, state)
		s.Require().NoError(err)
		s.assertSessionStarted(callbackResp, existing)

		s.Assert().Empty(s.consentsFor(existing))
	})

	s.Run("5. the login and signup gates still apply", func() {
		unknown := "noconsent-nologin@" + consentTestUserDomain
		_, err := s.startMailOTP(ctx, unknown, frontierv1beta1.FlowIntent_FLOW_INTENT_LOGIN, nil)
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeNotFound, connect.CodeOf(err))
		s.Assert().Equal(0, s.userCount(unknown))

		existing := "noconsent-unspecified@" + consentTestUserDomain
		_, err = s.startMailOTP(ctx, existing, frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP, nil)
		s.Require().Error(err)
		s.Assert().Equal(connect.CodeAlreadyExists, connect.CodeOf(err))
	})
}

func TestEndToEndConsentDisabledRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(ConsentDisabledRegressionTestSuite))
}
