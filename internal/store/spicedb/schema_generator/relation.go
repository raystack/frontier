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

	roleId := strings.ReplaceAll(str.DefaultStringIfEmpty(relation.Role.Id, relation.RoleId), "-", "_")
	roleNSId := str.DefaultStringIfEmpty(relation.Role.Namespace.Id, relation.Role.NamespaceId)
	if roleNSId != "" && roleNSId != transformedRelation.Resource.ObjectType {
		return &pb.Relationship{}, errors.New(fmt.Sprintf("Role %s doesnt exist in %s", roleId, transformedRelation.Resource.ObjectType))
	}

	transformedRelation.Relation = roleId
	return transformedRelation, nil
}

func transformObjectAndSubject(relation relation.Relation) (*pb.Relationship, error) {
	objectNSId := strings.ReplaceAll(str.DefaultStringIfEmpty(relation.ObjectNamespace.Id, relation.ObjectNamespaceId), "-", "_")
	subjectNSId := strings.ReplaceAll(str.DefaultStringIfEmpty(relation.SubjectNamespace.Id, relation.SubjectNamespaceId), "-", "_")

	return &pb.Relationship{
		Resource: &pb.ObjectReference{
			ObjectId:   relation.ObjectId,
			ObjectType: objectNSId,
		},
		Subject: &pb.SubjectReference{
			Object: &pb.ObjectReference{
				ObjectId:   relation.SubjectId,
				ObjectType: subjectNSId,
			},
			OptionalRelation: relation.SubjectRoleId,
		},
	}, nil
}

func TransformCheckRelation(relation relation.Relation) (*pb.Relationship, error) {
	return transformObjectAndSubject(relation)
}
