package schema_generator

import (
	"errors"
	"fmt"
	"strings"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/pkg/str"
)

func TransformRelation(relation relation.Relation) (*pb.Relationship, error) {
	transformedRelation, err := transformObjectAndSubject(relation)
	if err != nil {
		return nil, err
	}

	roleID := strings.ReplaceAll(str.DefaultStringIfEmpty(relation.Role.ID, relation.RoleID), "-", "_")
	roleNSID := str.DefaultStringIfEmpty(relation.Role.Namespace.ID, relation.Role.NamespaceID)
	if roleNSID != "" && roleNSID != transformedRelation.Resource.ObjectType {
		return &pb.Relationship{}, errors.New(fmt.Sprintf("Role %s doesnt exist in %s", roleID, transformedRelation.Resource.ObjectType))
	}

	transformedRelation.Relation = roleID
	return transformedRelation, nil
}

func transformObjectAndSubject(relation relation.Relation) (*pb.Relationship, error) {
	objectNSID := strings.ReplaceAll(str.DefaultStringIfEmpty(relation.ObjectNamespace.ID, relation.ObjectNamespaceID), "-", "_")
	subjectNSID := strings.ReplaceAll(str.DefaultStringIfEmpty(relation.SubjectNamespace.ID, relation.SubjectNamespaceID), "-", "_")

	return &pb.Relationship{
		Resource: &pb.ObjectReference{
			ObjectId:   relation.ObjectID,
			ObjectType: objectNSID,
		},
		Subject: &pb.SubjectReference{
			Object: &pb.ObjectReference{
				ObjectId:   relation.SubjectID,
				ObjectType: subjectNSID,
			},
			OptionalRelation: relation.SubjectRoleID,
		},
	}, nil
}

func TransformCheckRelation(relation relation.Relation) (*pb.Relationship, error) {
	return transformObjectAndSubject(relation)
}
