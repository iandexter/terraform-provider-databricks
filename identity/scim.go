package identity

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// URN is a custom type for the SCIM spec for the schema
type URN string

// Possible schema URNs for the Databricks SCIM api
const (
	UserSchema             URN = "urn:ietf:params:scim:schemas:core:2.0:User"
	ServicePrincipalSchema URN = "urn:ietf:params:scim:schemas:core:2.0:ServicePrincipal"
	WorkspaceUserSchema    URN = "urn:ietf:params:scim:schemas:extension:workspace:2.0:User"
	PatchOp                URN = "urn:ietf:params:scim:api:messages:2.0:PatchOp"
	GroupSchema            URN = "urn:ietf:params:scim:schemas:core:2.0:Group"
)

// GroupMember contains information of a member in a scim group
// TODO: merge GroupMember, GroupsListItem into valueItem
type GroupMember struct {
	Display string `json:"display,omitempty"`
	Value   string `json:"value,omitempty"`
	Ref     string `json:"$ref,omitempty"`

	// https://tools.ietf.org/html/rfc7643#page-64
	Type string `json:"type,omitempty"`
}

// GroupsListItem contains information about group item
// TODO: merge GroupMember, GroupsListItem into valueItem
type GroupsListItem struct {
	// TODO: combine entitlementsListItem & roleListItem into this one
	Display string `json:"display,omitempty"`
	Value   string `json:"value,omitempty"`

	// https://tools.ietf.org/html/rfc7643#page-64
	Type string `json:"type,omitempty"`
}

// TODO: merge GroupMember, GroupsListItem into valueItem
type valueItem struct {
	Value string `json:"value,omitempty"`
}

// ScimGroup contains information about the SCIM group
type ScimGroup struct {
	ID           string        `json:"id,omitempty"`
	Schemas      []URN         `json:"schemas,omitempty"`
	DisplayName  string        `json:"displayName,omitempty"`
	Members      []GroupMember `json:"members,omitempty"`
	Groups       []GroupMember `json:"groups,omitempty"`
	Roles        []valueItem   `json:"roles,omitempty"`
	Entitlements entitlements  `json:"entitlements,omitempty"`
}

// HasMember returns true if group has given user or another group id as member
func (g ScimGroup) HasMember(memberID string) bool {
	for _, member := range g.Members {
		if member.Value == memberID {
			return true
		}
	}
	return false
}

// HasRole returns true if group has a role
func (g ScimGroup) HasRole(role string) bool {
	for _, groupRole := range g.Roles {
		if groupRole.Value == role {
			return true
		}
	}
	return false
}

// GroupList contains a list of groups fetched from a list api call from SCIM api
type GroupList struct {
	TotalResults int32       `json:"totalResults,omitempty"`
	StartIndex   int32       `json:"startIndex,omitempty"`
	ItemsPerPage int32       `json:"itemsPerPage,omitempty"`
	Schemas      []URN       `json:"schemas,omitempty"`
	Resources    []ScimGroup `json:"resources,omitempty"`
}

var entitlementMapping = map[string]string{
	"allow-cluster-create":       "allow_cluster_create",
	"allow-instance-pool-create": "allow_instance_pool_create",
	"databricks-sql-access":      "allow_sql_analytics_access", // TODO: state change
}

// order is important for tests
var possibleEntitlements = []string{
	"allow-cluster-create",
	"allow-instance-pool-create",
	"databricks-sql-access"}

type entitlements []valueItem

func (e entitlements) readIntoData(d *schema.ResourceData) error {
	for _, ent := range e {
		field_name := entitlementMapping[ent.Value]
		if err := d.Set(field_name, true); err != nil {
			return err
		}
	}
	return nil
}

func readEntitlementsFromData(d *schema.ResourceData) entitlements {
	var e entitlements
	for _, entitlement := range possibleEntitlements {
		field_name := entitlementMapping[entitlement]
		if d.Get(field_name).(bool) {
			e = append(e, valueItem{
				Value: entitlement,
			})
		}
	}
	return e
}

func addEntitlementsToSchema(s *map[string]*schema.Schema) {
	for _, entitlement := range possibleEntitlements {
		field_name := entitlementMapping[entitlement]
		(*s)[field_name] = &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
		}
	}
}

type email struct {
	Type    interface{} `json:"type,omitempty"`
	Value   string      `json:"value,omitempty"`
	Primary interface{} `json:"primary,omitempty"`
}

// ScimUser is a struct that contains all the information about a SCIM user
type ScimUser struct {
	ID            string            `json:"id,omitempty"`
	Emails        []email           `json:"emails,omitempty"`
	DisplayName   string            `json:"displayName,omitempty" tf:"alias:display_name"`
	Active        bool              `json:"active,omitempty"`
	Schemas       []URN             `json:"schemas,omitempty"`
	UserName      string            `json:"userName,omitempty" tf:"alias:user_name"`
	ApplicationID string            `json:"applicationId,omitempty" tf:"alias:application_id"`
	Groups        []GroupsListItem  `json:"groups,omitempty"`
	Name          map[string]string `json:"name,omitempty"`
	Roles         []valueItem       `json:"roles,omitempty"`
	Entitlements  entitlements      `json:"entitlements,omitempty"`
}

// HasRole returns true if group has a role
func (u ScimUser) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r.Value == role {
			return true
		}
	}
	return false
}

// UserList contains a list of Users fetched from a list api call from SCIM api
type UserList struct {
	TotalResults int32      `json:"totalResults,omitempty"`
	StartIndex   int32      `json:"startIndex,omitempty"`
	ItemsPerPage int32      `json:"itemsPerPage,omitempty"`
	Schemas      []URN      `json:"schemas,omitempty"`
	Resources    []ScimUser `json:"resources,omitempty"`
}

type patchOperation struct {
	Op    string      `json:"op,omitempty"`
	Path  string      `json:"path,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type patchRequest struct {
	Schemas    []URN            `json:"schemas,omitempty"`
	Operations []patchOperation `json:"Operations,omitempty"`
}

func scimPatchRequest(op, path, value string) patchRequest {
	o := patchOperation{
		Op:   op,
		Path: path,
	}
	if value != "" {
		o.Value = []valueItem{{value}}
	}
	return patchRequest{
		Schemas:    []URN{PatchOp},
		Operations: []patchOperation{o},
	}
}
