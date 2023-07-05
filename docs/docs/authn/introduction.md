---
title: Overview
---
import Mermaid from '@theme/Mermaid';

# Overview

Authentication is the process of verifying the identity of a user. This is done by checking the user's credentials 
against a database of verified users. The database is populated with user credentials during the registration process.
If the credentials are valid, the user is granted access to the system. In some cases, the user's credentials are 
stored directly in the database like passwords, api keys. In other cases, the user's identity is verified by a third 
party like Google, Facebook. In this case, the user's credentials are stored in the third party's database and the 
user is redirected to the third party's login page to verify their identity.

Shield is an opinionated authentication server that provides a set of tools to help you implement authentication in
your application. It provides a couple of authentication strategies out of the box that you need to configure to start
using. What strategy you choose depends on your application's interface and the type of credentials you want to use.

:::note
A user is always authenticated to prove its identity before it can be authorized to access a
resource. This is why authentication is always the first step in the process of accessing a resource. 
:::
Authentication can be done by a human user or machine service user.

## User Authentication

For human users, Shield provides a couple of authentication strategies:

1. **Social login** - user's identity is verified by a third party like Google.
2. **Email one time password(OTP)** - user's identity is verified by sending a one time password to the user's email address.

Identity verification is the first step, but after it is verified, the system needs to remember that the human user is 
authenticated to avoid having to authenticate again for every request. HTTP is a stateless protocol, one way to do this 
over HTTP is to use cookies. Cookies are stored in the user's browser and sent with every request to the server. 
The server can then use the cookie to identify the user and verify if the session created using the cookie is 
still valid. If it is, the user no longer needs to authenticate again. This is how most web applications work. 

On high level here are some of the steps involved in the authentication process when working with a web application:

<Mermaid chart = {`sequenceDiagram
    %% participant User
    participant User
    participant Authentication Server
    participant Third-Party OIDC Provider
    participant Resource
   %
    User ->> Authentication Server: Sends request to access resource
      alt Check Authentication
        Authentication Server ->> Authentication Server: Check if user is authenticated
        alt User Not Authenticated
            Authentication Server -->> User: Redirects to login page
            User ->> Authentication Server: Selects third party provider (say Google)
            Authentication Server ->> Third-Party OIDC Provider: Redirects to OIDC login page
            User -->> Third-Party OIDC Provider: Enters credentials on OIDC login page
            Third-Party OIDC Provider ->> User: Sends authorization code
            User ->> Third-Party OIDC Provider: Exchanges code for access token and ID token
            Third-Party OIDC Provider -->> User: Retrieves access token and ID token
            User ->> Authentication Server: Sends ID token for verification
            Authentication Server ->> Third-Party OIDC Provider: Verifies ID token
            alt ID token valid
                Authentication Server -->> Resource: Continues to get requested resource
            else ID token invalid
                Authentication Server -->> User: Returns error
            end
        else User Authenticated
            Note over Authentication Server, Resource: Assuming user has permissions to access the resource
            Authentication Server ->> Resource: Continues to get the requested resource
            Resource ->> User: Return Requested Resource
        end
    end`}
/>

<br/>

Detailed Authentication Flow: 

1. A user sends a request to access a resource. The resource can be a web page, an API endpoint, a file, etc.
2. The request can either be intercepted by the authentication server or the frontend application can proxy the request
   to the authentication server. 
3. The authentication server checks if the user is authenticated by checking if the request contains a valid cookie.
4. If the user is not authenticated, the authentication server or the frontend redirects the user to the login page.
5. Login page can either contain a form to enter the user's credentials or a button to redirect the user to a third 
   party login page.
6. One the authentication flow is finished, the authentication server verifies the user's credentials.
7. After verification, if the user is accessing the authentication server for the first time a new user is created else
   the existing user is retrieved from the database. 
8. Finally, the authentication server redirects the user to the resource. (or the next step for validating authorization
can start)

## Service User Authentication

For machine service users, the authentication process is little different as the http requests won't be coming from a web
browser. Instead, the requests will be coming from a service like a external backup service or an SDK. In this case, the
service needs to send the service user's credentials with every request. This is done by sending the credentials in 
the request headers. The authentication server then verifies the credentials to check if the service user is authenticated.
If the credentials are valid, the resource grants access to the service user.
Shield provides a couple of authentication strategies for service users:

1. **Client ID/Secret** - service user's identity is verified by checking if the client id and secret are valid. Also known
   as client credentials grant or API grant.
2. **Private/Public Key JWT** - identity is verified by checking if the token sent in the headers is created and signed by a
private key that matches the public key stored in the database. Also known as JWT grant.

Shield can also return a token that the service can use to access the resource instead of using their credentials on
every request. This token is called a bearer access token. The service then sends the access token with every
request to the resource. Use of access token is optional, and it is short-lived.
