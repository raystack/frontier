# Frontier Connect RPC Sign-in Helper Script

A Python script to automate the authentication process using Frontier's Connect RPC APIs.

## Setup

1. Install dependencies:
```bash
pip install -r requirements.txt
```

## Usage

Basic usage with default settings (Connect RPC on port 8002):
```bash
python3 signin_helper.py your-email@example.com
```

With custom settings:
```bash
python3 signin_helper.py your-email@example.com \
  --base-url http://localhost:8002 \
  --db-host localhost \
  --db-port 5432 \
  --db-name frontier \
  --db-user frontier \
  --db-password frontier
```

Test the migrated ListUsers API after authentication:
```bash
python3 signin_helper.py your-email@example.com --test-list-users
```

Test with pagination to get ALL users:
```bash
python3 signin_helper.py your-email@example.com --test-list-users
```

Test without pagination (single page only):
```bash
python3 signin_helper.py your-email@example.com --test-list-users --no-pagination
```

Test the migrated CreateUser API after authentication:
```bash
python3 signin_helper.py your-email@example.com --test-create-user
```

Test the migrated GetUser API with a specific user ID:
```bash
python3 signin_helper.py your-email@example.com --test-get-user USER_ID_HERE
```

Test the migrated GetCurrentUser API after authentication:
```bash
python3 signin_helper.py your-email@example.com --test-get-current-user
```

Test the migrated UpdateUser API with a specific user ID:
```bash
python3 signin_helper.py your-email@example.com --test-update-user USER_ID_HERE
```

Test the migrated UpdateCurrentUser API after authentication:
```bash
python3 signin_helper.py your-email@example.com --test-update-current-user
```

Test the migrated EnableUser API with a specific user ID:
```bash
python3 signin_helper.py your-email@example.com --test-enable-user USER_ID_HERE
```

Test the migrated DisableUser API with a specific user ID:
```bash
python3 signin_helper.py your-email@example.com --test-disable-user USER_ID_HERE
```

Test the migrated DeleteUser API with a specific user ID:
```bash
python3 signin_helper.py your-email@example.com --test-delete-user USER_ID_HERE
```

Test the migrated ListUserGroups API with a specific user ID:
```bash
python3 signin_helper.py your-email@example.com --test-list-user-groups USER_ID_HERE
```

Test the migrated ListCurrentUserGroups API after authentication:
```bash
python3 signin_helper.py your-email@example.com --test-list-current-user-groups
```

Test the migrated ListOrganizationsByUser API with a specific user ID:
```bash
python3 signin_helper.py your-email@example.com --test-list-organizations-by-user USER_ID_HERE
```

Test the migrated ListOrganizationsByCurrentUser API after authentication:
```bash
python3 signin_helper.py your-email@example.com --test-list-organizations-by-current-user
```

Test the migrated ListProjectsByUser API with a specific user ID:
```bash
python3 signin_helper.py your-email@example.com --test-list-projects-by-user USER_ID_HERE
```

Test the migrated ListProjectsByCurrentUser API after authentication:
```bash
python3 signin_helper.py your-email@example.com --test-list-projects-by-current-user
```

Test the migrated ListServiceUsers API with a specific organization ID:
```bash
python3 signin_helper.py your-email@example.com --test-list-service-users ORG_ID_HERE
```

Test CreateUser then GetUser using the newly created user:
```bash
python3 signin_helper.py your-email@example.com --test-create-user --test-get-user created
```

Test CreateUser then UpdateUser using the newly created user:
```bash
python3 signin_helper.py your-email@example.com --test-create-user --test-update-user created
```

Test CreateUser then EnableUser using the newly created user:
```bash
python3 signin_helper.py your-email@example.com --test-create-user --test-enable-user created
```

Test CreateUser then DisableUser using the newly created user:
```bash
python3 signin_helper.py your-email@example.com --test-create-user --test-disable-user created
```

Test CreateUser then DeleteUser using the newly created user:
```bash
python3 signin_helper.py your-email@example.com --test-create-user --test-delete-user created
```

Test CreateUser then ListUserGroups using the newly created user:
```bash
python3 signin_helper.py your-email@example.com --test-create-user --test-list-user-groups created
```

Test CreateUser then ListOrganizationsByUser using the newly created user:
```bash
python3 signin_helper.py your-email@example.com --test-create-user --test-list-organizations-by-user created
```

Test CreateUser then ListProjectsByUser using the newly created user:
```bash
python3 signin_helper.py your-email@example.com --test-create-user --test-list-projects-by-user created
```

Test all APIs together:
```bash
python3 signin_helper.py your-email@example.com --test-list-users --test-create-user --test-get-user created --test-get-current-user --test-update-user created --test-update-current-user --test-enable-user created --test-disable-user created --test-delete-user created --test-list-user-groups created --test-list-current-user-groups --test-list-organizations-by-user created --test-list-organizations-by-current-user --test-list-projects-by-user created --test-list-projects-by-current-user
```

## What it does

1. **Start Auth Flow**: Calls the Connect RPC `Authenticate` endpoint with email and `mailotp` strategy
2. **Fetch OTP**: Queries the PostgreSQL `flows` table to get the generated OTP from the `nonce` field
3. **Complete Auth**: Uses the Connect RPC `AuthCallback` endpoint with OTP and flow ID
4. **Extract Session**: Prints authentication cookies and session headers for API usage
5. **Test APIs** (optional):
   - Tests the migrated `ListUsers` Connect RPC API with pagination to fetch ALL users (with `--test-list-users`)
   - Tests the migrated `CreateUser` Connect RPC API to create a test user (with `--test-create-user`)
   - Tests the migrated `GetUser` Connect RPC API to fetch a specific user by ID (with `--test-get-user`)
   - Tests the migrated `GetCurrentUser` Connect RPC API to fetch the current authenticated user (with `--test-get-current-user`)
   - Tests the migrated `UpdateUser` Connect RPC API to update a user by ID (with `--test-update-user`)
   - Tests the migrated `UpdateCurrentUser` Connect RPC API to update the current authenticated user (with `--test-update-current-user`)
   - Tests the migrated `EnableUser` Connect RPC API to enable a user by ID (with `--test-enable-user`)
   - Tests the migrated `DisableUser` Connect RPC API to disable a user by ID (with `--test-disable-user`)
   - Tests the migrated `DeleteUser` Connect RPC API to delete a user by ID (with `--test-delete-user`)
   - Tests the migrated `ListUserGroups` Connect RPC API to list groups for a user by ID (with `--test-list-user-groups`)
   - Tests the migrated `ListCurrentUserGroups` Connect RPC API to list groups for the current authenticated user (with `--test-list-current-user-groups`)
   - Tests the migrated `ListOrganizationsByUser` Connect RPC API to list organizations for a specific user by ID (with `--test-list-organizations-by-user`)
   - Tests the migrated `ListOrganizationsByCurrentUser` Connect RPC API to list organizations for the current authenticated user (with `--test-list-organizations-by-current-user`)
   - Tests the migrated `ListProjectsByUser` Connect RPC API to list projects for a specific user by ID (with `--test-list-projects-by-user`)
   - Tests the migrated `ListProjectsByCurrentUser` Connect RPC API to list projects for the current authenticated user (with `--test-list-projects-by-current-user`)

## Connect RPC Endpoints Used

- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/Authenticate`
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/AuthCallback`
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/ListUsers` (when `--test-list-users` is used)
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/CreateUser` (when `--test-create-user` is used)
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/GetUser` (when `--test-get-user` is used)
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/GetCurrentUser` (when `--test-get-current-user` is used)
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/UpdateUser` (when `--test-update-user` is used)
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/UpdateCurrentUser` (when `--test-update-current-user` is used)
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/EnableUser` (when `--test-enable-user` is used)
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/DisableUser` (when `--test-disable-user` is used)
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/DeleteUser` (when `--test-delete-user` is used)
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/ListUserGroups` (when `--test-list-user-groups` is used)
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/ListCurrentUserGroups` (when `--test-list-current-user-groups` is used)
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/ListOrganizationsByUser` (when `--test-list-organizations-by-user` is used)
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/ListOrganizationsByCurrentUser` (when `--test-list-organizations-by-current-user` is used)
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/ListProjectsByUser` (when `--test-list-projects-by-user` is used)
- **POST** `{base_url}/raystack.frontier.v1beta1.FrontierService/ListProjectsByCurrentUser` (when `--test-list-projects-by-current-user` is used)

## Output

The script will output:
- Authentication flow state and endpoint information
- OTP/code retrieved from database
- Authentication cookies and session headers
- Curl-ready Cookie header string

Example output:
```
🚀 Starting Connect RPC authentication flow for your-email@example.com
📧 Starting authentication flow for your-email@example.com
✅ Authentication flow started successfully
🔑 Flow State: abc123-def456-ghi789
🔍 Searching for OTP in database for your-email@example.com
🎯 Found OTP/Code: 123456
🔐 Completing authentication with code: 123456
✅ Authentication completed successfully
🔑 X-Frontier-Session-Id: session_abc123...

🍪 Authentication Cookies:
   frontier_session=abc123...

📋 Cookie Header for curl:
Cookie: frontier_session=abc123...

🎉 Authentication successful!
```

## Requirements

- Python 3.6+
- Access to Frontier PostgreSQL database
- Frontier Connect RPC server running (default: localhost:8002)
- Email-based OTP authentication enabled in Frontier

## Database Schema

The script works with the actual `flows` table schema:
- `id` - Flow identifier (UUID)
- `method` - Authentication method used
- `email` - User email
- `nonce` - Nonce value (may contain OTP)
- `metadata` - JSONB metadata (primary location for OTP/code)
- `created_at` - Creation timestamp
- `expires_at` - Expiration timestamp

The script looks for OTP/code in:
1. `nonce` field (if it's numeric and 4+ digits)
2. `metadata` field under keys: `otp`, `code`, `token`, or `verification_code`