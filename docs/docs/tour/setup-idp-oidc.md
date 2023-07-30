# Setup an external IDP for OIDC

This tour page provides instructions on how to configure and use [OIDC (OpenID Connect)](../concepts/glossary.md#oidc) for authentication via an external [Identity Provider (IDP)](../concepts/glossary.md#identity-providers-idps). OIDC is an authentication protocol that allows applications to verify the identity of users based on the authentication performed by an IDP.

### Pre-requisites

Before proceeding with the configuration, make sure you have the following:

- Access to an external IDP that supports OIDC authentication. In this tour, we will use Google as the example IDP.
- [Configure the server](../reference/configurations.md#server-configurations) to listen on the desired domain and port. For example `http://localhost:8000`

## Configuration Steps

Follow the steps below to configure OIDC authentication via an external IDP:

1. **Set up an OIDC callback URL**: Determine the URL where the external IDP will redirect users after authentication. For example: **`http://localhost:8000/v1beta1/auth/callback`** . Replace localhost:8000 with the appropriate domain and port for your application. Ensure that this URL is accessible by your application and can receive requests. <br/><br/>**Note:** For the purpose of this tutorial, we have setup [`examples/authn`](https://github.com/raystack/frontier/tree/main/examples/authn) folder to test the enitre flow. The [redirect URL](../concepts/glossary.md#redirect-uri) will remain the same as the above example **`http://localhost:8000/v1beta1/auth/callback`**

2. **Configure the OIDC authentication server:** Obtain the necessary configuration details from your external IDP. For Google, follow these steps:

   - Create a project in the Google Cloud Console (if you haven't already).
   - Enable the Google Identity Platform API for your project.
   - Create OAuth 2.0 credentials and obtain the [Client ID](../concepts/glossary.md#client-id) and [Client Secret](../concepts/glossary.md#client-secret).
   - Set the Authorized Redirect URI to the OIDC callback URL you determined in step 1.
   - Note the [issuer URL](../concepts/glossary.md#issuer-url), which for Google is "https://accounts.google.com".

3. **Update the frontier configuration file:**

   - Open the Frontier server configuration file that handles authentication.
   - Add the following OIDC-related configurations under **`app.authentication`** section:

   ```yaml
   callback_host: http://localhost:8000/v1beta1/auth/callback
   oidc_config:
     google:
       client_id: "xxxxx.apps.googleusercontent.com"
       client_secret: "xxxxx"
       issuer_url: "https://accounts.google.com"
   ```

   - Replace **xxxxx.apps.googleusercontent.com** with your Google Client ID.
   - Replace **xxxxx** with your Google Client Secret.
   - Ensure that **callback_host** matches the **callback URL** you determined in step 1.
   - Update **issuer_url** if you're using a different IDP.

   :::tip Tip
   Each external IDP may have its own specific configuration requirements. Consult the documentation of your chosen IDP for detailed instructions on how to configure OIDC authentication with Frontier.
   :::

## Generate the RSA keys

Generate 2 RSA keys for Auth token generation using the following command and add the file **path** and **iss(issuer)** in the server config in the app_authentication section like this.

```bash
$ ./frontier server keygen > ./temp/rsa
```

```yaml
token:
  rsa_path: ./temp/rsa
  iss: "http://localhost.frontier"
```

:::info Additional Info

<details>
<summary>RSA key pairs and issuer</summary>
Once authenticated, the Frontier server responds with a JWT token having user context. <br/><br/>

**RSA Key Pair:**
OIDC relies on cryptographic mechanisms to sign and verify tokens, such as JWTs (JSON Web Tokens).
The RSA key pair consists of a private key for token signing and a corresponding public key for token verification.
The private key is securely stored by the authorization server or identity provider (IDP) and is used to generate the digital signature on the tokens.
The public key is made available to clients and relying parties to verify the authenticity and integrity of the tokens.
In this configuration, the rsa_path parameter specifies the location of the RSA key files used for token generation.

The **issuer URL** uniquely identifies the IDP or authorization server that issues the tokens.

By configuring the RSA key path and issuer URL, Frontier can generate tokens with appropriate signatures and metadata, allowing services/applications to securely verify and authenticate the tokens received from the Frontier server after user authentication.

</details>
:::
