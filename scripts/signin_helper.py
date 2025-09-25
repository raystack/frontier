#!/usr/bin/env python3
"""
Frontier Sign-in Helper Script (Connect RPC)

This script automates the sign-in process using Connect RPC APIs by:
1. Starting authentication flow with email via Connect RPC API
2. Fetching OTP/code from the flow table in PostgreSQL
3. Completing authentication callback and retrieving the session cookie

Usage: python signin_helper.py <email>
"""

import sys
import time
import json
import requests
import psycopg2
from psycopg2.extras import RealDictCursor
import argparse
from datetime import datetime, timedelta

class FrontierSignInHelper:
    def __init__(self, base_url="http://localhost:8002", db_config=None):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })
        self.db_config = db_config or {
            'host': 'localhost',
            'port': 5432,
            'database': 'frontier',
            'user': 'frontier',
            'password': 'frontier'
        }

    def list_auth_strategies(self):
        """List available authentication strategies"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/ListAuthStrategies"

        payload = {}

        print(f"🔍 Fetching available authentication strategies")

        try:
            response = self.session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            print(f"✅ Available strategies:")

            strategies = data.get('strategies', [])
            for strategy in strategies:
                print(f"   - {strategy.get('name')}")

            return strategies

        except requests.RequestException as e:
            print(f"❌ Error fetching strategies: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response: {e.response.text}")
            return []

    def initiate_signin(self, email, strategy="mailotp"):
        """Initiate authentication flow with email using Connect RPC"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/Authenticate"

        payload = {
            "strategyName": strategy,  # Use specified strategy for authentication
            "email": email,
            "redirectOnstart": False
        }

        print(f"📧 Starting authentication flow for {email}")

        try:
            response = self.session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            print(f"✅ Authentication flow started successfully")
            print(f"📝 Response: {json.dumps(data, indent=2)}")

            # Extract state from response (this should be the flow UUID)
            state = data.get('state')
            endpoint = data.get('endpoint')

            if state:
                print(f"🔑 Flow ID/State: {state}")
                print(f"🔗 Endpoint: {endpoint}")
                return state, endpoint
            else:
                print("⚠️  No state/flow ID found in response")
                return None, None

        except requests.RequestException as e:
            print(f"❌ Error starting authentication flow: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response: {e.response.text}")
            return None, None

    def get_otp_from_db(self, flow_id):
        """Fetch OTP/code from the flow table using the specific flow ID"""
        print(f"🔍 Fetching OTP from database using flow ID: {flow_id}")

        try:
            conn = psycopg2.connect(**self.db_config)
            cursor = conn.cursor(cursor_factory=RealDictCursor)

            # Query the specific flow by ID
            query = """
            SELECT id, method, email, nonce, metadata, created_at, expires_at
            FROM flows
            WHERE id = %s
            """
            cursor.execute(query, (flow_id,))

            row = cursor.fetchone()

            if row:
                print(f"📋 Found flow record:")
                print(f"🗂️  Flow ID: {row['id']}")
                print(f"🔧 Method: {row.get('method')}")
                print(f"📧 Email: {row.get('email')}")
                print(f"🔑 Nonce: {row.get('nonce')}")
                print(f"📅 Created: {row['created_at']}")
                print(f"⏰ Expires: {row['expires_at']}")

                # The nonce field contains the OTP
                nonce = row.get('nonce')
                if nonce:
                    print(f"🎯 Using nonce as OTP: {nonce}")
                    conn.close()
                    return str(nonce)
                else:
                    conn.close()
                    print("❌ No nonce (OTP) found in this flow record")
                    return None
            else:
                conn.close()
                print(f"❌ Flow record not found for ID: {flow_id}")
                return None

        except Exception as e:
            print(f"❌ Database error: {e}")
            return None

    def complete_signin(self, email, code, state, strategy="mailotp"):
        """Complete authentication using Connect RPC AuthCallback"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/AuthCallback"

        payload = {
            "strategyName": strategy,
            "code": code,
            "state": state
        }

        print(f"🔐 Completing authentication with code: {code}")

        try:
            response = self.session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            print(f"✅ Authentication completed successfully")
            print(f"📝 Response: {json.dumps(data, indent=2)}")

            # Extract session information from headers
            session_info = {}
            headers = response.headers

            # Look for session-related headers
            for header_name, header_value in headers.items():
                if 'session' in header_name.lower() or 'cookie' in header_name.lower():
                    session_info[header_name] = header_value
                    print(f"🔑 {header_name}: {header_value}")

            # Extract cookies from response
            cookies = {}
            for cookie in self.session.cookies:
                cookies[cookie.name] = cookie.value

            if cookies:
                print(f"🍪 Authentication Cookies:")
                for name, value in cookies.items():
                    print(f"   {name}={value}")

                # Print curl-friendly format
                cookie_header = "; ".join([f"{name}={value}" for name, value in cookies.items()])
                print(f"\n📋 Cookie Header for curl:")
                print(f"Cookie: {cookie_header}")

                return cookies, session_info
            else:
                print("⚠️  No cookies received")
                # Still return session info if available
                return None, session_info

        except requests.RequestException as e:
            print(f"❌ Error completing authentication: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return None, None

    def signin_flow(self, email, strategy="mailotp"):
        """Complete Connect RPC authentication flow"""
        print(f"🚀 Starting Connect RPC authentication flow for {email}")
        print("=" * 50)

        # Step 0: List available strategies (for debugging)
        print("🔧 Debug: Checking available authentication strategies...")
        self.list_auth_strategies()
        print()

        # Step 1: Start authentication flow
        flow_id, endpoint = self.initiate_signin(email, strategy)

        if not flow_id:
            print("❌ Could not start authentication flow")
            return None, None

        # Step 2: Get OTP/code from database using the flow ID
        code = self.get_otp_from_db(flow_id)

        if not code:
            print("❌ Could not retrieve OTP/code")
            return None, None

        # Step 3: Complete authentication
        cookies, session_info = self.complete_signin(email, code, flow_id, strategy)

        if cookies or session_info:
            print("\n🎉 Authentication successful!")
            return cookies, session_info
        else:
            print("\n❌ Authentication failed")
            return None, None

    def test_list_users_api(self, cookies, use_pagination=True):
        """Test the migrated ListUsers Connect RPC API with authentication and pagination"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/ListUsers"

        print(f"🧪 Testing ListUsers Connect RPC API with pagination")
        print(f"🔗 URL: {url}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        # Add cookies to the session
        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        all_users = []
        page_num = 1
        page_size = 10  # Use smaller page size to test pagination
        total_pages_fetched = 0
        max_pages = 50  # Safety limit to prevent infinite loops

        try:
            while page_num <= max_pages:
                payload = {
                    "pageSize": page_size,
                    "pageNum": page_num,
                    "keyword": "",
                    "orgId": "",
                    "groupId": "",
                    "state": ""
                }

                print(f"📄 Fetching page {page_num} (page size: {page_size})")

                response = api_session.post(url, json=payload)
                response.raise_for_status()

                data = response.json()
                users = data.get('users', [])
                count = data.get('count', 0)

                print(f"   ✅ Page {page_num}: Got {len(users)} users (total in response: {count})")

                if not users:
                    print(f"   🏁 No more users found on page {page_num}")
                    break

                # Add users to our collection
                all_users.extend(users)
                total_pages_fetched += 1

                # If we got fewer users than page_size, we've reached the end
                if len(users) < page_size:
                    print(f"   🏁 Reached end of results (got {len(users)} < {page_size})")
                    break

                # If not using pagination, just fetch one page
                if not use_pagination:
                    break

                page_num += 1

            # Summary
            print(f"\n📊 PAGINATION SUMMARY:")
            print(f"   📄 Pages fetched: {total_pages_fetched}")
            print(f"   👥 Total users collected: {len(all_users)}")
            print(f"   📏 Page size used: {page_size}")

            # Show first few and last few users
            print(f"\n👥 USER SAMPLE:")
            for i, user in enumerate(all_users[:5], 1):
                print(f"   {i}. {user.get('name')} ({user.get('email')}) - ID: {user.get('id', 'N/A')}")

            if len(all_users) > 10:
                print(f"   ... ({len(all_users) - 10} users in between) ...")
                for i, user in enumerate(all_users[-5:], len(all_users) - 4):
                    print(f"   {i}. {user.get('name')} ({user.get('email')}) - ID: {user.get('id', 'N/A')}")
            elif len(all_users) > 5:
                for i, user in enumerate(all_users[5:], 6):
                    print(f"   {i}. {user.get('name')} ({user.get('email')}) - ID: {user.get('id', 'N/A')}")

            return True

        except requests.RequestException as e:
            print(f"❌ Error calling ListUsers API: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return False

    def test_create_user_api(self, cookies):
        """Test the migrated CreateUser Connect RPC API with authentication"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/CreateUser"

        print(f"🧪 Testing CreateUser Connect RPC API")
        print(f"🔗 URL: {url}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        # Add cookies to the session
        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        # Generate a unique test user
        import random
        test_suffix = random.randint(1000, 9999)
        test_email = f"testuser{test_suffix}@example.com"

        payload = {
            "body": {
                "email": test_email,
                "name": f"test-user-{test_suffix}",
                "title": f"Test User {test_suffix}",
                "metadata": {
                    "source": "api_test",
                    "test_run": True
                }
            }
        }

        print(f"👤 Creating test user: {test_email}")

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            user = data.get('user', {})

            print(f"   ✅ User created successfully!")
            print(f"   🆔 User ID: {user.get('id', 'N/A')}")
            print(f"   📧 Email: {user.get('email', 'N/A')}")
            print(f"   👤 Name: {user.get('name', 'N/A')}")
            print(f"   📝 Title: {user.get('title', 'N/A')}")

            if user.get('metadata'):
                print(f"   🏷️  Metadata: {user.get('metadata')}")

            if user.get('createdAt'):
                print(f"   📅 Created: {user.get('createdAt')}")

            return True, user

        except requests.RequestException as e:
            print(f"❌ Error calling CreateUser API: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return False, None

    def test_get_user_api(self, cookies, user_id, silent=False):
        """Test the migrated GetUser Connect RPC API with authentication"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/GetUser"

        if not silent:
            print(f"🧪 Testing GetUser Connect RPC API")
            print(f"🔗 URL: {url}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        # Add cookies to the session
        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        payload = {
            "id": user_id
        }

        if not silent:
            print(f"👤 Fetching user with ID: {user_id}")

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            user = data.get('user', {})

            if not silent:
                print(f"   ✅ User fetched successfully!")
                print(f"   🆔 User ID: {user.get('id', 'N/A')}")
                print(f"   📧 Email: {user.get('email', 'N/A')}")
                print(f"   👤 Name: {user.get('name', 'N/A')}")
                print(f"   📝 Title: {user.get('title', 'N/A')}")

                if user.get('metadata'):
                    print(f"   🏷️  Metadata: {user.get('metadata')}")

                if user.get('createdAt'):
                    print(f"   📅 Created: {user.get('createdAt')}")

                if user.get('updatedAt'):
                    print(f"   🔄 Updated: {user.get('updatedAt')}")

            return True, user

        except requests.RequestException as e:
            if not silent:
                print(f"❌ Error calling GetUser API: {e}")
                if hasattr(e, 'response') and e.response:
                    print(f"Response status: {e.response.status_code}")
                    print(f"Response: {e.response.text}")
            return False, None

    def test_get_current_user_api(self, cookies, silent=False):
        """Test the migrated GetCurrentUser Connect RPC API with authentication"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/GetCurrentUser"

        if not silent:
            print(f"🧪 Testing GetCurrentUser Connect RPC API")
            print(f"🔗 URL: {url}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        # Add cookies to the session
        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        payload = {}  # GetCurrentUser doesn't need any parameters

        if not silent:
            print(f"👤 Fetching current authenticated user")

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            user = data.get('user')
            service_user = data.get('serviceuser')

            if user:
                if not silent:
                    print(f"   ✅ Current user fetched successfully!")
                    print(f"   🆔 User ID: {user.get('id', 'N/A')}")
                    print(f"   📧 Email: {user.get('email', 'N/A')}")
                    print(f"   👤 Name: {user.get('name', 'N/A')}")
                    print(f"   📝 Title: {user.get('title', 'N/A')}")

                    if user.get('metadata'):
                        print(f"   🏷️  Metadata: {user.get('metadata')}")

                    if user.get('createdAt'):
                        print(f"   📅 Created: {user.get('createdAt')}")

                return True, user
            elif service_user:
                if not silent:
                    print(f"   ✅ Current service user fetched successfully!")
                    print(f"   🆔 Service User ID: {service_user.get('id', 'N/A')}")
                    print(f"   📧 Email: {service_user.get('email', 'N/A')}")
                    print(f"   👤 Title: {service_user.get('title', 'N/A')}")

                    if service_user.get('metadata'):
                        print(f"   🏷️  Metadata: {service_user.get('metadata')}")

                return True, service_user
            else:
                if not silent:
                    print(f"   ⚠️  No user or service user found in response")
                return False, None

        except requests.RequestException as e:
            if not silent:
                print(f"❌ Error calling GetCurrentUser API: {e}")
                if hasattr(e, 'response') and e.response:
                    print(f"Response status: {e.response.status_code}")
                    print(f"Response: {e.response.text}")
            return False, None

    def test_update_user_api(self, cookies, user_id=None, user_data=None):
        """Test the migrated UpdateUser Connect RPC API with authentication"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/UpdateUser"

        print(f"🧪 Testing UpdateUser Connect RPC API")
        print(f"🔗 URL: {url}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        # Add cookies to the session
        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        # Use provided user_id or generate a test ID
        test_user_id = user_id if user_id else "test-user-id-for-update"

        # First, get the current user data to show "before" state
        print(f"📋 Fetching current user data for comparison...")
        get_user_success, original_user = self.test_get_user_api(cookies, test_user_id, silent=True)

        if not get_user_success:
            print(f"⚠️  Could not fetch current user data - proceeding with update anyway")
            original_user = None

        # Use provided user data or create default test data
        if user_data is None:
            user_data = {
                "title": "Updated User via Connect RPC",
                "email": "updated-user@test.com",
                "name": "updated-user-slug",
                "avatar": "updated-avatar.jpg",
                "metadata": {
                    "department": "engineering",
                    "role": "senior-developer",
                    "updated": True
                }
            }

        payload = {
            "id": test_user_id,
            "body": user_data
        }

        print(f"🔄 Updating user with ID: {test_user_id}")

        # Show before/after comparison if we have original data
        if original_user:
            print(f"\n📊 BEFORE → AFTER COMPARISON:")
            print(f"   📝 Title: '{original_user.get('title', 'N/A')}' → '{user_data.get('title', 'N/A')}'")
            print(f"   📧 Email: '{original_user.get('email', 'N/A')}' → '{user_data.get('email', 'N/A')}'")
            print(f"   👤 Name: '{original_user.get('name', 'N/A')}' → '{user_data.get('name', 'N/A')}'")
            print(f"   🖼️  Avatar: '{original_user.get('avatar', 'N/A')}' → '{user_data.get('avatar', 'N/A')}'")

            # Compare metadata
            original_metadata = original_user.get('metadata', {})
            new_metadata = user_data.get('metadata', {})
            print(f"   🏷️  Metadata changes:")

            # Show all keys from both original and new metadata
            all_keys = set(list(original_metadata.keys()) + list(new_metadata.keys()))
            for key in sorted(all_keys):
                old_val = original_metadata.get(key, '[not set]')
                new_val = new_metadata.get(key, '[not set]')
                if old_val != new_val:
                    print(f"      {key}: '{old_val}' → '{new_val}'")
                else:
                    print(f"      {key}: '{old_val}' (unchanged)")
        else:
            print(f"📝 Update data: {user_data}")

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            updated_user = data.get('user')

            print(f"\n   ✅ User updated successfully!")
            print(f"   🆔 User ID: {updated_user.get('id', 'N/A')}")
            print(f"   📧 Email: {updated_user.get('email', 'N/A')}")
            print(f"   👤 Name: {updated_user.get('name', 'N/A')}")
            print(f"   📝 Title: {updated_user.get('title', 'N/A')}")
            print(f"   🖼️ Avatar: {updated_user.get('avatar', 'N/A')}")

            if updated_user.get('metadata'):
                print(f"   🏷️  Metadata: {updated_user.get('metadata')}")

            if updated_user.get('createdAt'):
                print(f"   📅 Created: {updated_user.get('createdAt')}")

            if updated_user.get('updatedAt'):
                print(f"   🔄 Updated: {updated_user.get('updatedAt')}")

            return True, updated_user

        except requests.RequestException as e:
            print(f"❌ Error calling UpdateUser API: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return False, None

    def test_update_current_user_api(self, cookies, user_data=None):
        """Test the migrated UpdateCurrentUser Connect RPC API with authentication"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/UpdateCurrentUser"

        print(f"🧪 Testing UpdateCurrentUser Connect RPC API")
        print(f"🔗 URL: {url}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        # Add cookies to the session
        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        # First, get the current user data to show "before" state and get the email
        print(f"📋 Fetching current user data for comparison...")
        current_user_success, current_user = self.test_get_current_user_api(cookies, silent=True)

        if not current_user_success:
            print(f"⚠️  Could not fetch current user data - cannot proceed with UpdateCurrentUser")
            return False, None

        current_email = current_user.get('email')
        if not current_email:
            print(f"⚠️  Could not get current user email - cannot proceed with UpdateCurrentUser")
            return False, None

        # Use provided user data or create default test data
        if user_data is None:
            user_data = {
                "title": "Updated Current User via Connect RPC",
                "email": current_email,  # Must match current user's email
                "name": "updated-current-user-slug",
                "avatar": "updated-current-avatar.jpg",
                "metadata": {
                    "department": "product",
                    "role": "product-manager",
                    "updated_via": "connect_rpc",
                    "timestamp": "2024-01-01"
                }
            }
        else:
            # Ensure email matches current user
            user_data["email"] = current_email

        payload = {"body": user_data}

        print(f"🔄 Updating current authenticated user")

        # Show before/after comparison
        if current_user:
            print(f"\n📊 BEFORE → AFTER COMPARISON:")
            print(f"   📝 Title: '{current_user.get('title', 'N/A')}' → '{user_data.get('title', 'N/A')}'")
            print(f"   📧 Email: '{current_user.get('email', 'N/A')}' → '{user_data.get('email', 'N/A')}' (must match)")
            print(f"   👤 Name: '{current_user.get('name', 'N/A')}' → '{user_data.get('name', 'N/A')}'")
            print(f"   🖼️  Avatar: '{current_user.get('avatar', 'N/A')}' → '{user_data.get('avatar', 'N/A')}'")

            # Compare metadata
            current_metadata = current_user.get('metadata', {})
            new_metadata = user_data.get('metadata', {})
            print(f"   🏷️  Metadata changes:")

            # Show all keys from both current and new metadata
            all_keys = set(list(current_metadata.keys()) + list(new_metadata.keys()))
            for key in sorted(all_keys):
                old_val = current_metadata.get(key, '[not set]')
                new_val = new_metadata.get(key, '[not set]')
                if old_val != new_val:
                    print(f"      {key}: '{old_val}' → '{new_val}'")
                else:
                    print(f"      {key}: '{old_val}' (unchanged)")

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            updated_user = data.get('user')

            print(f"\n   ✅ Current user updated successfully!")
            print(f"   🆔 User ID: {updated_user.get('id', 'N/A')}")
            print(f"   📧 Email: {updated_user.get('email', 'N/A')}")
            print(f"   👤 Name: {updated_user.get('name', 'N/A')}")
            print(f"   📝 Title: {updated_user.get('title', 'N/A')}")
            print(f"   🖼️ Avatar: {updated_user.get('avatar', 'N/A')}")

            if updated_user.get('metadata'):
                print(f"   🏷️  Metadata: {updated_user.get('metadata')}")

            if updated_user.get('createdAt'):
                print(f"   📅 Created: {updated_user.get('createdAt')}")

            if updated_user.get('updatedAt'):
                print(f"   🔄 Updated: {updated_user.get('updatedAt')}")

            return True, updated_user

        except requests.RequestException as e:
            print(f"❌ Error calling UpdateCurrentUser API: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return False, None

    def test_enable_user_api(self, cookies, user_id):
        """Test the migrated EnableUser Connect RPC API with authentication"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/EnableUser"

        print(f"🔗 URL: {url}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        # Add cookies to the session
        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        payload = {
            "id": user_id
        }

        print(f"🔓 Enabling user: {user_id}")

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            print(f"   ✅ User enabled successfully!")
            print(f"   🆔 User ID: {user_id}")

            return True

        except requests.RequestException as e:
            print(f"❌ Error calling EnableUser API: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return False

    def test_disable_user_api(self, cookies, user_id):
        """Test the migrated DisableUser Connect RPC API with authentication"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/DisableUser"

        print(f"🔗 URL: {url}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        # Add cookies to the session
        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        payload = {
            "id": user_id
        }

        print(f"🔒 Disabling user: {user_id}")

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            print(f"   ✅ User disabled successfully!")
            print(f"   🆔 User ID: {user_id}")

            return True

        except requests.RequestException as e:
            print(f"❌ Error calling DisableUser API: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return False

    def test_delete_user_api(self, cookies, user_id):
        """Test the migrated DeleteUser Connect RPC API with authentication"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/DeleteUser"

        print(f"🧪 Testing DeleteUser Connect RPC API")
        print(f"🔗 URL: {url}")
        print(f"🆔 User ID: {user_id}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        # Set cookies if provided
        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        # Prepare the request payload
        payload = {
            "id": user_id
        }

        print(f"📦 Request payload: {payload}")

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()

            print(f"\n   ✅ User deleted successfully!")
            print(f"   🆔 Deleted User ID: {user_id}")
            print(f"   📋 Response: {data}")

            return True

        except requests.RequestException as e:
            print(f"❌ Error calling DeleteUser API: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return False

    def test_list_user_groups_api(self, cookies, user_id, org_id=None):
        """Test the migrated ListUserGroups Connect RPC API with authentication"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/ListUserGroups"

        print(f"🧪 Testing ListUserGroups Connect RPC API")
        print(f"🔗 URL: {url}")
        print(f"🆔 User ID: {user_id}")
        if org_id:
            print(f"🏢 Organization ID: {org_id}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        # Set cookies if provided
        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        # Prepare the request payload
        payload = {
            "id": user_id
        }
        if org_id:
            payload["orgId"] = org_id

        print(f"📦 Request payload: {payload}")

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            groups = data.get('groups', [])

            print(f"\n   ✅ ListUserGroups API call successful!")
            print(f"   📊 Found {len(groups)} groups for user")

            if groups:
                print(f"   📋 Groups:")
                for i, group in enumerate(groups, 1):
                    print(f"      {i}. 🆔 ID: {group.get('id', 'N/A')}")
                    print(f"         📛 Name: {group.get('name', 'N/A')}")
                    print(f"         📝 Title: {group.get('title', 'N/A')}")
                    print(f"         🏢 Organization ID: {group.get('orgId', 'N/A')}")
                    print(f"         👥 Members: {group.get('membersCount', 0)}")
                    if group.get('createdAt'):
                        print(f"         📅 Created: {group.get('createdAt')}")
                    if group.get('updatedAt'):
                        print(f"         🔄 Updated: {group.get('updatedAt')}")
                    print()
            else:
                print(f"   ℹ️  User is not a member of any groups")

            return True, groups

        except requests.RequestException as e:
            print(f"❌ Error calling ListUserGroups API: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return False, []

    def test_list_current_user_groups_api(self, cookies, org_id=None, with_permissions=None):
        """Test the migrated ListCurrentUserGroups Connect RPC API with authentication"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/ListCurrentUserGroups"
        print(f"🧪 Testing ListCurrentUserGroups Connect RPC API")
        print(f"🔗 URL: {url}")
        if org_id:
            print(f"🏢 Organization ID: {org_id}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        # Set cookies if provided
        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        # Prepare the request payload
        payload = {}
        if org_id:
            payload["orgId"] = org_id
        if with_permissions:
            payload["withPermissions"] = with_permissions

        print(f"📦 Request payload: {payload}")

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            groups = data.get('groups', [])
            access_pairs = data.get('accessPairs', [])

            print(f"\n   ✅ ListCurrentUserGroups API call successful!")
            print(f"   📊 Found {len(groups)} groups for current user")

            if groups:
                print(f"   📋 Groups:")
                for i, group in enumerate(groups, 1):
                    print(f"      {i}. 🆔 ID: {group.get('id', 'N/A')}")
                    print(f"         📛 Name: {group.get('name', 'N/A')}")
                    print(f"         📝 Title: {group.get('title', 'N/A')}")
                    print(f"         🏢 Organization ID: {group.get('orgId', 'N/A')}")
                    print(f"         👥 Members: {group.get('membersCount', 0)}")
                    if group.get('createdAt'):
                        print(f"         📅 Created: {group.get('createdAt')}")
                    if group.get('updatedAt'):
                        print(f"         🔄 Updated: {group.get('updatedAt')}")
                    print()
            else:
                print(f"   ℹ️  Current user is not a member of any groups")

            if access_pairs:
                print(f"   🔑 Access Pairs (Permissions):")
                for i, pair in enumerate(access_pairs, 1):
                    print(f"      {i}. 🆔 Group ID: {pair.get('groupId', 'N/A')}")
                    permissions = pair.get('permissions', [])
                    if permissions:
                        print(f"         🎯 Permissions: {', '.join(permissions)}")
                    else:
                        print(f"         🎯 Permissions: None")
                    print()
            elif with_permissions:
                print(f"   ℹ️  No permissions found for specified groups")

            return True, groups, access_pairs

        except requests.RequestException as e:
            print(f"❌ Error calling ListCurrentUserGroups API: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return False, [], []

    def test_list_organizations_by_user_api(self, cookies, user_id, state=None):
        """Test the migrated ListOrganizationsByUser Connect RPC API with authentication"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/ListOrganizationsByUser"
        print(f"🧪 Testing ListOrganizationsByUser Connect RPC API")
        print(f"🔗 URL: {url}")
        print(f"🆔 User ID: {user_id}")
        if state:
            print(f"📊 State Filter: {state}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        # Set cookies if provided
        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        # Prepare the request payload
        payload = {
            "id": user_id
        }
        if state:
            payload["state"] = state

        print(f"📦 Request payload: {payload}")

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            organizations = data.get('organizations', [])
            joinable_via_domain = data.get('joinableViaDomain', [])

            print(f"\n   ✅ ListOrganizationsByUser API call successful!")
            print(f"   📊 Found {len(organizations)} organizations for user")
            print(f"   🌐 Found {len(joinable_via_domain)} joinable organizations via domain")

            if organizations:
                print(f"   🏢 User Organizations:")
                for i, org in enumerate(organizations, 1):
                    print(f"      {i}. 🆔 ID: {org.get('id', 'N/A')}")
                    print(f"         📛 Name: {org.get('name', 'N/A')}")
                    print(f"         📝 Title: {org.get('title', 'N/A')}")
                    print(f"         📊 State: {org.get('state', 'N/A')}")
                    if org.get('createdAt'):
                        print(f"         📅 Created: {org.get('createdAt')}")
                    if org.get('updatedAt'):
                        print(f"         🔄 Updated: {org.get('updatedAt')}")
                    print()
            else:
                print(f"   ℹ️  User is not a member of any organizations")

            if joinable_via_domain:
                print(f"   🌐 Organizations Joinable via Email Domain:")
                for i, org in enumerate(joinable_via_domain, 1):
                    print(f"      {i}. 🆔 ID: {org.get('id', 'N/A')}")
                    print(f"         📛 Name: {org.get('name', 'N/A')}")
                    print(f"         📝 Title: {org.get('title', 'N/A')}")
                    print(f"         📊 State: {org.get('state', 'N/A')}")
                    print()
            else:
                print(f"   ℹ️  No organizations joinable via user's email domain")

            return True, organizations, joinable_via_domain

        except requests.RequestException as e:
            print(f"❌ Error calling ListOrganizationsByUser API: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return False, [], []

    def test_list_organizations_by_current_user_api(self, cookies, state=None):
        """Test the migrated ListOrganizationsByCurrentUser Connect RPC API with authentication"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/ListOrganizationsByCurrentUser"
        print(f"🧪 Testing ListOrganizationsByCurrentUser Connect RPC API")
        print(f"🔗 URL: {url}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        # Prepare the request payload
        payload = {}
        if state:
            payload["state"] = state
            print(f"🎯 State filter: {state}")

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            organizations = data.get('organizations', [])
            joinable_via_domain = data.get('joinableViaDomain', [])

            print(f"\n   ✅ ListOrganizationsByCurrentUser API call successful!")
            print(f"   📊 Found {len(organizations)} organizations for current user")
            print(f"   🌐 Found {len(joinable_via_domain)} joinable organizations via domain")

            if organizations:
                print(f"   🏢 User Organizations:")
                for i, org in enumerate(organizations, 1):
                    print(f"      {i}. 🆔 ID: {org.get('id', 'N/A')}")
                    print(f"         📛 Name: {org.get('name', 'N/A')}")
                    print(f"         📝 Title: {org.get('title', 'N/A')}")
                    print(f"         📊 State: {org.get('state', 'N/A')}")
                    if org.get('createdAt'):
                        print(f"         📅 Created: {org.get('createdAt')}")
                    print()  # Empty line for readability
            else:
                print(f"   ℹ️  No organizations found for current user")

            if joinable_via_domain:
                print(f"   🌐 Joinable Organizations via Email Domain:")
                for i, org in enumerate(joinable_via_domain, 1):
                    print(f"      {i}. 🆔 ID: {org.get('id', 'N/A')}")
                    print(f"         📛 Name: {org.get('name', 'N/A')}")
                    print(f"         📝 Title: {org.get('title', 'N/A')}")
                    print(f"         📊 State: {org.get('state', 'N/A')}")
                    if org.get('createdAt'):
                        print(f"         📅 Created: {org.get('createdAt')}")
                    print()  # Empty line for readability
            else:
                print(f"   ℹ️  No organizations joinable via current user's email domain")

            return True, organizations, joinable_via_domain

        except requests.RequestException as e:
            print(f"❌ Error calling ListOrganizationsByCurrentUser API: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return False, [], []

    def test_list_projects_by_user_api(self, cookies, user_id):
        """Test the migrated ListProjectsByUser Connect RPC API with authentication"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/ListProjectsByUser"
        print(f"🧪 Testing ListProjectsByUser Connect RPC API")
        print(f"🔗 URL: {url}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        # Prepare the request payload
        payload = {
            "id": user_id
        }

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            projects = data.get('projects', [])

            print(f"\n   ✅ ListProjectsByUser API call successful!")
            print(f"   📊 Found {len(projects)} projects for user")

            if projects:
                print(f"   🏗️  User Projects:")
                for i, project in enumerate(projects, 1):
                    print(f"      {i}. 🆔 ID: {project.get('id', 'N/A')}")
                    print(f"         📛 Name: {project.get('name', 'N/A')}")
                    print(f"         📝 Title: {project.get('title', 'N/A')}")
                    print(f"         🏢 Organization ID: {project.get('orgId', 'N/A')}")
                    print(f"         👥 Members Count: {project.get('membersCount', 'N/A')}")
                    if project.get('createdAt'):
                        print(f"         📅 Created: {project.get('createdAt')}")
                    print()  # Empty line for readability
            else:
                print(f"   ℹ️  No projects found for user")

            return True, projects

        except requests.RequestException as e:
            print(f"❌ Error calling ListProjectsByUser API: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return False, []

    def test_list_projects_by_current_user_api(self, cookies, org_id=None, with_member_count=False, non_inherited=False):
        """Test the migrated ListProjectsByCurrentUser Connect RPC API with authentication"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/ListProjectsByCurrentUser"
        print(f"🧪 Testing ListProjectsByCurrentUser Connect RPC API")
        print(f"🔗 URL: {url}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        # Prepare the request payload
        payload = {}
        if org_id:
            payload["orgId"] = org_id
            print(f"🏢 Organization filter: {org_id}")
        if with_member_count:
            payload["withMemberCount"] = with_member_count
            print(f"👥 Include member count: {with_member_count}")
        if non_inherited:
            payload["nonInherited"] = non_inherited
            print(f"🚫 Non-inherited only: {non_inherited}")

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            projects = data.get('projects', [])
            access_pairs = data.get('accessPairs', [])
            count = data.get('count', 0)

            print(f"\n   ✅ ListProjectsByCurrentUser API call successful!")
            print(f"   📊 Found {len(projects)} projects for current user (Total count: {count})")

            if projects:
                print(f"   🏗️  Current User Projects:")
                for i, project in enumerate(projects, 1):
                    print(f"      {i}. 🆔 ID: {project.get('id', 'N/A')}")
                    print(f"         📛 Name: {project.get('name', 'N/A')}")
                    print(f"         📝 Title: {project.get('title', 'N/A')}")
                    print(f"         🏢 Organization ID: {project.get('orgId', 'N/A')}")
                    print(f"         👥 Members Count: {project.get('membersCount', 'N/A')}")
                    if project.get('createdAt'):
                        print(f"         📅 Created: {project.get('createdAt')}")
                    print()  # Empty line for readability
            else:
                print(f"   ℹ️  No projects found for current user")

            if access_pairs:
                print(f"   🔐 Access Pairs:")
                for i, pair in enumerate(access_pairs, 1):
                    print(f"      {i}. 🆔 Project ID: {pair.get('projectId', 'N/A')}")
                    print(f"         🔑 Permissions: {', '.join(pair.get('permissions', []))}")
                    print()

            return True, projects, access_pairs

        except requests.RequestException as e:
            print(f"❌ Error calling ListProjectsByCurrentUser API: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return False, [], []

    def test_list_service_users_api(self, cookies, org_id, state=None):
        """Test the migrated ListServiceUsers Connect RPC API"""
        url = f"{self.base_url}/raystack.frontier.v1beta1.FrontierService/ListServiceUsers"
        print(f"🧪 Testing ListServiceUsers Connect RPC API")
        print(f"🔗 URL: {url}")

        # Create a new session with cookies for this API call
        api_session = requests.Session()
        api_session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

        # Add cookies to the session
        if cookies:
            for name, value in cookies.items():
                api_session.cookies.set(name, value)

        payload = {
            "orgId": org_id
        }

        if state:
            payload["state"] = state

        try:
            response = api_session.post(url, json=payload)
            response.raise_for_status()

            data = response.json()
            service_users = data.get('serviceusers', [])

            print(f"✅ ListServiceUsers API called successfully!")
            print(f"📊 Total Service Users: {len(service_users)}")

            if service_users:
                print("📋 Service Users:")
                for i, su in enumerate(service_users[:5], 1):  # Show first 5
                    print(f"   {i}. ID: {su.get('id', 'N/A')}")
                    print(f"      Title: {su.get('title', 'N/A')}")
                    print(f"      Org ID: {su.get('orgId', 'N/A')}")
                    print(f"      State: {su.get('state', 'N/A')}")
                    if su.get('metadata'):
                        print(f"      Metadata: {su.get('metadata')}")
                    print()

                if len(service_users) > 5:
                    print(f"   ... and {len(service_users) - 5} more service users")
            else:
                print("   ℹ️  No service users found")

            return True, service_users

        except requests.RequestException as e:
            print(f"❌ Error calling ListServiceUsers API: {e}")
            if hasattr(e, 'response') and e.response:
                print(f"Response status: {e.response.status_code}")
                print(f"Response: {e.response.text}")
            return False, []

def main():
    parser = argparse.ArgumentParser(description='Frontier Connect RPC Sign-in Helper')
    parser.add_argument('email', help='Email address to sign in with')
    parser.add_argument('--base-url', default='http://localhost:8002',
                       help='Base URL of Frontier Connect RPC API (default: http://localhost:8002)')
    parser.add_argument('--db-host', default='localhost', help='Database host')
    parser.add_argument('--db-port', type=int, default=5432, help='Database port')
    parser.add_argument('--db-name', default='frontier', help='Database name')
    parser.add_argument('--db-user', default='frontier', help='Database user')
    parser.add_argument('--db-password', default='frontier', help='Database password')
    parser.add_argument('--strategy', default='mailotp', help='Authentication strategy (default: mailotp)')
    parser.add_argument('--test-list-users', action='store_true', help='Test ListUsers API after authentication')
    parser.add_argument('--test-create-user', action='store_true', help='Test CreateUser API after authentication')
    parser.add_argument('--test-get-user', help='Test GetUser API with specific user ID after authentication')
    parser.add_argument('--test-get-current-user', action='store_true', help='Test GetCurrentUser API after authentication')
    parser.add_argument('--test-update-user', help='Test UpdateUser API with specific user ID after authentication')
    parser.add_argument('--test-update-current-user', action='store_true', help='Test UpdateCurrentUser API after authentication')
    parser.add_argument('--test-enable-user', help='Test EnableUser API with specific user ID after authentication')
    parser.add_argument('--test-disable-user', help='Test DisableUser API with specific user ID after authentication')
    parser.add_argument('--test-delete-user', help='Test DeleteUser API with specific user ID after authentication')
    parser.add_argument('--test-list-user-groups', help='Test ListUserGroups API with specific user ID after authentication')
    parser.add_argument('--test-list-current-user-groups', action='store_true', help='Test ListCurrentUserGroups API after authentication')
    parser.add_argument('--test-list-organizations-by-user', help='Test ListOrganizationsByUser API with specific user ID after authentication')
    parser.add_argument('--test-list-organizations-by-current-user', action='store_true', help='Test ListOrganizationsByCurrentUser API after authentication')
    parser.add_argument('--test-list-projects-by-user', help='Test ListProjectsByUser API with specific user ID after authentication')
    parser.add_argument('--test-list-projects-by-current-user', action='store_true', help='Test ListProjectsByCurrentUser API after authentication')
    parser.add_argument('--test-list-service-users', help='Test ListServiceUsers API with specific org ID after authentication')
    parser.add_argument('--no-pagination', action='store_true', help='Disable pagination (fetch only first page)')

    args = parser.parse_args()

    # Setup database configuration
    db_config = {
        'host': args.db_host,
        'port': args.db_port,
        'database': args.db_name,
        'user': args.db_user,
        'password': args.db_password
    }

    # Create helper instance
    helper = FrontierSignInHelper(args.base_url, db_config)

    # Run authentication flow
    cookies, session_info = helper.signin_flow(args.email, args.strategy)

    if cookies or session_info:
        print(f"\n🎯 Authentication completed for {args.email}")

        # Print session info summary
        if session_info:
            print(f"\n📋 Session Information:")
            for key, value in session_info.items():
                print(f"   {key}: {value}")

        # Test ListUsers API if requested
        if args.test_list_users and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED LISTUSERS API")
            print("=" * 50)

            use_pagination = not args.no_pagination
            api_success = helper.test_list_users_api(cookies, use_pagination)
            if api_success:
                print(f"\n🎉 ListUsers API test successful!")
            else:
                print(f"\n💥 ListUsers API test failed!")

        # Test CreateUser API if requested
        created_user_id = None
        if args.test_create_user and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED CREATEUSER API")
            print("=" * 50)

            create_success, created_user = helper.test_create_user_api(cookies)
            if create_success:
                print(f"\n🎉 CreateUser API test successful!")
                if created_user:
                    created_user_id = created_user.get('id')
                    print(f"📊 Created user details:")
                    print(f"   ID: {created_user_id}")
                    print(f"   Email: {created_user.get('email')}")
                    print(f"   Name: {created_user.get('name')}")
            else:
                print(f"\n💥 CreateUser API test failed!")

        # Test GetUser API if requested
        if args.test_get_user and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED GETUSER API")
            print("=" * 50)

            # Use the provided user ID or the created user ID if available
            user_id_to_fetch = args.test_get_user
            if user_id_to_fetch == "created" and created_user_id:
                user_id_to_fetch = created_user_id
                print(f"📋 Using newly created user ID for testing")

            get_success, fetched_user = helper.test_get_user_api(cookies, user_id_to_fetch)
            if get_success:
                print(f"\n🎉 GetUser API test successful!")
                if fetched_user:
                    print(f"📊 Fetched user details:")
                    print(f"   ID: {fetched_user.get('id')}")
                    print(f"   Email: {fetched_user.get('email')}")
                    print(f"   Name: {fetched_user.get('name')}")
            else:
                print(f"\n💥 GetUser API test failed!")

        # Test GetCurrentUser API if requested
        if args.test_get_current_user and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED GETCURRENTUSER API")
            print("=" * 50)

            current_user_success, current_user = helper.test_get_current_user_api(cookies)
            if current_user_success:
                print(f"\n🎉 GetCurrentUser API test successful!")
                if current_user:
                    print(f"📊 Current user details:")
                    print(f"   ID: {current_user.get('id')}")
                    print(f"   Email: {current_user.get('email')}")
                    user_type = "Service User" if 'serviceuser' in str(type(current_user)) else "User"
                    print(f"   Type: {user_type}")
            else:
                print(f"\n💥 GetCurrentUser API test failed!")

        # Test UpdateUser API if requested
        if args.test_update_user and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED UPDATEUSER API")
            print("=" * 50)

            test_user_id = args.test_update_user

            # Handle special case where user wants to update created user
            if test_user_id == "created" and created_user_id:
                test_user_id = created_user_id
                print(f"🔗 Using newly created user ID: {test_user_id}")
            elif test_user_id == "created" and not created_user_id:
                print(f"❌ Cannot use 'created' user ID - no user was created in this session")
                print(f"💡 Use --test-create-user flag first, or provide a specific user ID")
            else:
                print(f"🔗 Using provided user ID: {test_user_id}")

            if test_user_id != "created" or created_user_id:
                update_success, updated_user = helper.test_update_user_api(cookies, test_user_id)
                if update_success:
                    print(f"\n🎉 UpdateUser API test successful!")
                    if updated_user:
                        print(f"📊 Updated user details:")
                        print(f"   ID: {updated_user.get('id')}")
                        print(f"   Email: {updated_user.get('email')}")
                        print(f"   Name: {updated_user.get('name')}")
                        print(f"   Title: {updated_user.get('title')}")
                        print(f"   Avatar: {updated_user.get('avatar')}")
                else:
                    print(f"\n💥 UpdateUser API test failed!")

        # Test UpdateCurrentUser API if requested
        if args.test_update_current_user and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED UPDATECURRENTUSER API")
            print("=" * 50)

            update_success, updated_user = helper.test_update_current_user_api(cookies)
            if update_success:
                print(f"\n🎉 UpdateCurrentUser API test successful!")
                if updated_user:
                    print(f"📊 Updated current user details:")
                    print(f"   ID: {updated_user.get('id')}")
                    print(f"   Email: {updated_user.get('email')}")
                    print(f"   Name: {updated_user.get('name')}")
                    print(f"   Title: {updated_user.get('title')}")
                    print(f"   Avatar: {updated_user.get('avatar')}")
            else:
                print(f"\n💥 UpdateCurrentUser API test failed!")

        # Test EnableUser API if requested
        if args.test_enable_user and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED ENABLEUSER API")
            print("=" * 50)

            test_user_id = args.test_enable_user

            # Handle special case where user wants to enable created user
            if test_user_id == "created" and created_user_id:
                test_user_id = created_user_id
                print(f"🔗 Using newly created user ID: {test_user_id}")
            elif test_user_id == "created" and not created_user_id:
                print(f"❌ Cannot use 'created' user ID - no user was created in this session")
                print(f"💡 Use --test-create-user flag first, or provide a specific user ID")
            else:
                print(f"🔗 Using provided user ID: {test_user_id}")

            if test_user_id != "created" or created_user_id:
                enable_success = helper.test_enable_user_api(cookies, test_user_id)
                if enable_success:
                    print(f"\n🎉 EnableUser API test successful!")
                else:
                    print(f"\n💥 EnableUser API test failed!")

        # Test DisableUser API if requested
        if args.test_disable_user and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED DISABLEUSER API")
            print("=" * 50)

            test_user_id = args.test_disable_user

            # Handle special case where user wants to disable created user
            if test_user_id == "created" and created_user_id:
                test_user_id = created_user_id
                print(f"🔗 Using newly created user ID: {test_user_id}")
            elif test_user_id == "created" and not created_user_id:
                print(f"❌ Cannot use 'created' user ID - no user was created in this session")
                print(f"💡 Use --test-create-user flag first, or provide a specific user ID")
            else:
                print(f"🔗 Using provided user ID: {test_user_id}")

            if test_user_id != "created" or created_user_id:
                disable_success = helper.test_disable_user_api(cookies, test_user_id)
                if disable_success:
                    print(f"\n🎉 DisableUser API test successful!")
                else:
                    print(f"\n💥 DisableUser API test failed!")

        # Test DeleteUser API if requested
        if args.test_delete_user and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED DELETEUSER API")
            print("=" * 50)

            test_user_id = args.test_delete_user

            # Handle special case where user wants to delete created user
            if test_user_id == "created" and created_user_id:
                test_user_id = created_user_id
                print(f"🔗 Using newly created user ID: {test_user_id}")
            elif test_user_id == "created" and not created_user_id:
                print(f"❌ Cannot use 'created' user ID - no user was created in this session")
                print(f"💡 Use --test-create-user flag first, or provide a specific user ID")
            else:
                print(f"🔗 Using provided user ID: {test_user_id}")

            # Only proceed if we have a valid user ID to delete
            if test_user_id != "created" or created_user_id:
                delete_success = helper.test_delete_user_api(cookies, test_user_id)
                if delete_success:
                    print(f"\n🎉 DeleteUser API test successful!")
                else:
                    print(f"\n💥 DeleteUser API test failed!")

        # Test ListUserGroups API if requested
        if args.test_list_user_groups and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED LISTUSERGROUPS API")
            print("=" * 50)

            test_user_id = args.test_list_user_groups

            # Handle special case where user wants to list groups for created user
            if test_user_id == "created" and created_user_id:
                test_user_id = created_user_id
                print(f"🔗 Using newly created user ID: {test_user_id}")
            elif test_user_id == "created" and not created_user_id:
                print(f"❌ Cannot use 'created' user ID - no user was created in this session")
                print(f"💡 Use --test-create-user flag first, or provide a specific user ID")
            else:
                print(f"🔗 Using provided user ID: {test_user_id}")

            # Only proceed if we have a valid user ID to list groups for
            if test_user_id != "created" or created_user_id:
                groups_success, groups = helper.test_list_user_groups_api(cookies, test_user_id)
                if groups_success:
                    print(f"\n🎉 ListUserGroups API test successful!")
                else:
                    print(f"\n💥 ListUserGroups API test failed!")

        # Test ListCurrentUserGroups API if requested
        if args.test_list_current_user_groups and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED LISTCURRENTUSERGROUPS API")
            print("=" * 50)

            groups_success, groups, access_pairs = helper.test_list_current_user_groups_api(cookies)
            if groups_success:
                print(f"\n🎉 ListCurrentUserGroups API test successful!")
            else:
                print(f"\n💥 ListCurrentUserGroups API test failed!")

        # Test ListOrganizationsByUser API if requested
        if args.test_list_organizations_by_user and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED LISTORGANIZATIONSBYUSER API")
            print("=" * 50)

            test_user_id = args.test_list_organizations_by_user
            # Handle special case where user wants to list organizations for created user
            if test_user_id == "created" and created_user_id:
                test_user_id = created_user_id
                print(f"🔗 Using newly created user ID: {test_user_id}")
            elif test_user_id == "created" and not created_user_id:
                print(f"❌ Cannot use 'created' user ID - no user was created in this session")
                print(f"💡 Use --test-create-user flag first, or provide a specific user ID")
            else:
                print(f"🔗 Using provided user ID: {test_user_id}")

            # Only proceed if we have a valid user ID to list organizations for
            if test_user_id != "created" or created_user_id:
                orgs_success, organizations, joinable_orgs = helper.test_list_organizations_by_user_api(cookies, test_user_id)
                if orgs_success:
                    print(f"\n🎉 ListOrganizationsByUser API test successful!")
                else:
                    print(f"\n💥 ListOrganizationsByUser API test failed!")

        # Test ListOrganizationsByCurrentUser API if requested
        if args.test_list_organizations_by_current_user and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED LISTORGANIZATIONSBYCURRENTUSER API")
            print("=" * 50)

            orgs_success, organizations, joinable_orgs = helper.test_list_organizations_by_current_user_api(cookies)
            if orgs_success:
                print(f"\n🎉 ListOrganizationsByCurrentUser API test successful!")
            else:
                print(f"\n💥 ListOrganizationsByCurrentUser API test failed!")

        # Test ListProjectsByUser API if requested
        if args.test_list_projects_by_user and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED LISTPROJECTSBYUSER API")
            print("=" * 50)

            test_user_id = args.test_list_projects_by_user

            # Handle special case where user wants to list projects for created user
            if test_user_id == "created" and created_user_id:
                test_user_id = created_user_id
                print(f"🔗 Using newly created user ID: {test_user_id}")
            elif test_user_id == "created" and not created_user_id:
                print(f"❌ Cannot use 'created' user ID - no user was created in this session")
                print(f"💡 Use --test-create-user flag first, or provide a specific user ID")
            else:
                print(f"🔗 Using provided user ID: {test_user_id}")

            # Only proceed if we have a valid user ID to list projects for
            if test_user_id != "created" or created_user_id:
                projects_success, projects = helper.test_list_projects_by_user_api(cookies, test_user_id)
                if projects_success:
                    print(f"\n🎉 ListProjectsByUser API test successful!")
                else:
                    print(f"\n💥 ListProjectsByUser API test failed!")

        # Test ListProjectsByCurrentUser API if requested
        if args.test_list_projects_by_current_user and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED LISTPROJECTSBYCURRENTUSER API")
            print("=" * 50)

            projects_success, projects, access_pairs = helper.test_list_projects_by_current_user_api(cookies)
            if projects_success:
                print(f"\n🎉 ListProjectsByCurrentUser API test successful!")
            else:
                print(f"\n💥 ListProjectsByCurrentUser API test failed!")

        # Test ListServiceUsers API if requested
        if args.test_list_service_users and cookies:
            print(f"\n" + "=" * 50)
            print("🧪 TESTING MIGRATED LISTSERVICEUSERS API")
            print("=" * 50)
            service_users_success, service_users = helper.test_list_service_users_api(cookies, args.test_list_service_users)
            if service_users_success:
                print(f"\n🎉 ListServiceUsers API test successful!")
            else:
                print(f"\n💥 ListServiceUsers API test failed!")

        sys.exit(0)
    else:
        print(f"\n💥 Authentication failed for {args.email}")
        sys.exit(1)

if __name__ == "__main__":
    main()