package constants

type Role string
type Permission string

const (
	RoleUser       Role = "user"
	RoleAdmin      Role = "admin"
	RoleSuperAdmin Role = "super_admin"
	RoleModerator  Role = "moderator"
)

const (
	PermReadProducts   Permission = "products:read"
	PermWriteProducts  Permission = "products:write"
	PermDeleteProducts Permission = "products:delete"
	PermReadCategories Permission = "categories:read"
	PermWriteCategories Permission = "categories:write"
	PermDeleteCategories Permission = "categories:delete"
	PermReadUsers      Permission = "users:read"
	PermWriteUsers     Permission = "users:write"
	PermDeleteUsers    Permission = "users:delete"
	PermReadAnalytics  Permission = "analytics:read"
)

var RolePermissions = map[Role][]Permission{
	RoleUser: {
		PermReadProducts,
	},
	RoleModerator: {
		PermReadProducts, PermWriteProducts,
		PermReadCategories,
		PermReadUsers,
	},
	RoleAdmin: {
		PermReadProducts, PermWriteProducts, PermDeleteProducts,
		PermReadCategories, PermWriteCategories, PermDeleteCategories,
		PermReadUsers, PermWriteUsers,
		PermReadAnalytics,
	},
	RoleSuperAdmin: {
		PermReadProducts, PermWriteProducts, PermDeleteProducts,
		PermReadCategories, PermWriteCategories, PermDeleteCategories,
		PermReadUsers, PermWriteUsers, PermDeleteUsers,
		PermReadAnalytics,
	},
}
