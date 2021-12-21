package schema_generator

import (
	"errors"
	"fmt"
	"strings"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/odpf/shield/model"
	"github.com/odpf/shield/pkg/utils"
)

func TransformRelation(relation model.Relation) (*pb.Relationship, error) {
	roleId := strings.ReplaceAll(utils.DefaultStringIfEmpty(relation.Role.Id, relation.RoleId), "-", "_")
	objectNSId := strings.ReplaceAll(utils.DefaultStringIfEmpty(relation.ObjectNamespace.Id, relation.ObjectNamespaceId), "-", "_")
	subjectNSId := strings.ReplaceAll(utils.DefaultStringIfEmpty(relation.SubjectNamespace.Id, relation.SubjectNamespaceId), "-", "_")

	roleNSId := utils.DefaultStringIfEmpty(relation.Role.Namespace.Id, relation.Role.NamespaceId)

	if roleNSId != "" && roleNSId != objectNSId {
		return &pb.Relationship{}, errors.New(fmt.Sprintf("Role %s doesnt exist in %s", roleId, objectNSId))
	}

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
		Relation: roleId,
	}, nil
}
