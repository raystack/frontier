package e2e_test

import (
	"context"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/frontier/test/e2e/testbench"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *APIRegressionTestSuite) TestUpdateUserMetadata() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)

	newMetadata := func(fields map[string]any) *structpb.Struct {
		metadata, err := structpb.NewStruct(fields)
		s.Require().NoError(err)
		return metadata
	}

	getMetadata := func(userID string) map[string]any {
		resp, err := s.testBench.Client.GetUser(ctxAdmin, connect.NewRequest(&frontierv1beta1.GetUserRequest{Id: userID}))
		s.Require().NoError(err)
		return resp.Msg.GetUser().GetMetadata().AsMap()
	}

	createUser := func(name, email string) string {
		resp, err := s.testBench.Client.CreateUser(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title:    "user metadata update",
				Email:    email,
				Name:     name,
				Metadata: newMetadata(map[string]any{"description": "core user", "size": "10"}),
			},
		}))
		s.Require().NoError(err)
		return resp.Msg.GetUser().GetId()
	}

	s.Run("update user by id", func() {
		email := "user-metadata-update@raystack.org"
		userID := createUser("user_metadata_update", email)

		s.Run("1. update without metadata keeps the stored metadata", func() {
			resp, err := s.testBench.Client.UpdateUser(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateUserRequest{
				Id: userID,
				Body: &frontierv1beta1.UserRequestBody{
					Title: "user metadata update renamed",
					Email: email,
				},
			}))
			s.Require().NoError(err)
			s.Assert().Equal("user metadata update renamed", resp.Msg.GetUser().GetTitle())
			s.Assert().Equal(map[string]any{"description": "core user", "size": "10"}, getMetadata(userID))
		})

		s.Run("2. update with metadata replaces the stored metadata", func() {
			_, err := s.testBench.Client.UpdateUser(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateUserRequest{
				Id: userID,
				Body: &frontierv1beta1.UserRequestBody{
					Title:    "user metadata update renamed",
					Email:    email,
					Metadata: newMetadata(map[string]any{"size": "20"}),
				},
			}))
			s.Require().NoError(err)
			s.Assert().Equal(map[string]any{"size": "20"}, getMetadata(userID))
		})

		s.Run("3. update with empty metadata clears the stored metadata", func() {
			_, err := s.testBench.Client.UpdateUser(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateUserRequest{
				Id: userID,
				Body: &frontierv1beta1.UserRequestBody{
					Title:    "user metadata update renamed",
					Email:    email,
					Metadata: &structpb.Struct{},
				},
			}))
			s.Require().NoError(err)
			s.Assert().Empty(getMetadata(userID))
		})
	})

	s.Run("update current user", func() {
		email := "current-user-metadata-update@raystack.org"
		userID := createUser("current_user_metadata_update", email)

		userCookie, err := testbench.AuthenticateUser(context.Background(), s.testBench.Client, email)
		s.Require().NoError(err)
		ctxUser := testbench.ContextWithAuth(context.Background(), userCookie)

		s.Run("1. update without metadata keeps the stored metadata", func() {
			resp, err := s.testBench.Client.UpdateCurrentUser(ctxUser, connect.NewRequest(&frontierv1beta1.UpdateCurrentUserRequest{
				Body: &frontierv1beta1.UserRequestBody{
					Title: "current user metadata update renamed",
					Email: email,
				},
			}))
			s.Require().NoError(err)
			s.Assert().Equal("current user metadata update renamed", resp.Msg.GetUser().GetTitle())
			s.Assert().Equal(map[string]any{"description": "core user", "size": "10"}, getMetadata(userID))
		})

		s.Run("2. update with metadata replaces the stored metadata", func() {
			_, err := s.testBench.Client.UpdateCurrentUser(ctxUser, connect.NewRequest(&frontierv1beta1.UpdateCurrentUserRequest{
				Body: &frontierv1beta1.UserRequestBody{
					Title:    "current user metadata update renamed",
					Email:    email,
					Metadata: newMetadata(map[string]any{"size": "20"}),
				},
			}))
			s.Require().NoError(err)
			s.Assert().Equal(map[string]any{"size": "20"}, getMetadata(userID))
		})

		s.Run("3. update with empty metadata clears the stored metadata", func() {
			_, err := s.testBench.Client.UpdateCurrentUser(ctxUser, connect.NewRequest(&frontierv1beta1.UpdateCurrentUserRequest{
				Body: &frontierv1beta1.UserRequestBody{
					Title:    "current user metadata update renamed",
					Email:    email,
					Metadata: &structpb.Struct{},
				},
			}))
			s.Require().NoError(err)
			s.Assert().Empty(getMetadata(userID))
		})
	})
}
