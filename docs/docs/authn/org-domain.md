# Domain Whitelisting

In Frontier adding a user to an Organization can be done in either of these ways:
- **By allowing all users from a Organization's trusted domains**: Say `raystack.org` is a trusted domain for `raystack` Organization, then any user with email address `*@raystack.org` can join the Organization without explicit invitation.
- **By explicitely inviting individual users to the Organization**: Assuming a user with a public domain email address `@gmail.com` is required to be part of the `raystack` Organization, then the user can join the Organization only after accepting the invitation for the same.

---

## How It Works !!

Whitelisting trusted domains in Frontier allows you to explicitly specify which domains are considered safe for user to join a specific organization. An Organization member with appropriate permissions can add or remove the trsuted domains for that particular Organization. However, to prevent someone else to use your domain to sign up in Frontier Organization, domain ownership needs to be proved. 

Every domain (like `raystack.org`) has a set of DNS records that can be viewed by anyone on the internet. DNS records tell computers how to find your website and where to deliver your company's email messages. 

To verify domain ownership, Frontier provides a verification token which needs to be added to the DNS records of the domain. Once the token is added, Frontier can perform a DNS lookup to check if the verification code matches the DNS record. If the verification code matches, the domain is verified and the domain verification status is updated to `verified`.

---

## Domain Ownership Verification - Step by Step

1. **Create a request to add a domain to an Org**: Organization Admins or members with update permissions at Org level can Add domain in Frontier using this API.

2. **Access Domain Registrar**: Log in to your domain registrar's control panel or website. The domain registrar is the company where you purchased your domain (e.g., GoDaddy, Namecheap, etc.).

3. **Locate DNS Records Settings**: In the domain registrar's control panel, locate the section related to DNS records settings or DNS management.

4. **Add Verification Code to DNS Records**:  Once you're in the DNS management section, look for the option to add new DNS records. Choose the record type "TXT" (Text), which allows you to add free-form text information to the domain's DNS records.

Frontier verification token looks something like this `_frontier-challenge:1234567890123456`. The token is a combination of a prefix `_frontier-challenge:` and a random string of 16 characters. The prefix is used to identify the token as a Frontier verification token. The random string is used to ensure that the token is unique and not guessable.

5. **Paste Verification code**: Copy a verification code which the above Frontier API returns. This is the same record Frontier expects in the DNS record of the domain an Organization claims to own.

6. **Save Changes**: Save the new TXT *record in your domain's DNS settings. The record might take some time (typically a few minutes to an hour) to propagate across the internet. These TXT records can be removed once verified.

7. **Verify Ownership**: After adding the TXT record, Frontier Verify Org Domain API can be used to perform a DNS lookup to check if the verification code matches the DNS record. 

:::info
Until the domain verification status is marked **`pending`**, Frontier won't consider the domain to be a trusted source and won't allow users to join the Org unless they're invite explicitely. 

The org domain must be verified within 7 days of adding it to the trusted domains list. Failing which the entire process of adding and verifying the domain needs to be repeated with a new validation token.
:::
