package domain

import "time"

// User represents a better-auth user record.
type User struct {
	ID            string    `db:"id"             json:"id"`
	Name          string    `db:"name"           json:"name"`
	Email         string    `db:"email"          json:"email"`
	EmailVerified bool      `db:"email_verified" json:"emailVerified"`
	Image         *string   `db:"image"          json:"image"`
	CreatedAt     time.Time `db:"created_at"     json:"createdAt"`
	UpdatedAt     time.Time `db:"updated_at"     json:"updatedAt"`
}

// Session represents a better-auth session record.
type Session struct {
	ID        string    `db:"id"         json:"id"`
	ExpiresAt time.Time `db:"expires_at" json:"expiresAt"`
	Token     string    `db:"token"      json:"token"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
	IPAddress *string   `db:"ip_address" json:"ipAddress"`
	UserAgent *string   `db:"user_agent" json:"userAgent"`
	UserID    string    `db:"user_id"    json:"userId"`
}

// Account represents a better-auth OAuth/credential account.
type Account struct {
	ID                     string     `db:"id"                        json:"id"`
	AccountID              string     `db:"account_id"                json:"accountId"`
	ProviderID             string     `db:"provider_id"               json:"providerId"`
	UserID                 string     `db:"user_id"                   json:"userId"`
	AccessToken            *string    `db:"access_token"              json:"accessToken"`
	RefreshToken           *string    `db:"refresh_token"             json:"refreshToken"`
	IDToken                *string    `db:"id_token"                  json:"idToken"`
	AccessTokenExpiresAt   *time.Time `db:"access_token_expires_at"   json:"accessTokenExpiresAt"`
	RefreshTokenExpiresAt  *time.Time `db:"refresh_token_expires_at"  json:"refreshTokenExpiresAt"`
	Scope                  *string    `db:"scope"                     json:"scope"`
	Password               *string    `db:"password"                  json:"password"`
	CreatedAt              time.Time  `db:"created_at"                json:"createdAt"`
	UpdatedAt              time.Time  `db:"updated_at"                json:"updatedAt"`
}

// Verification represents a better-auth email/phone verification record.
type Verification struct {
	ID         string     `db:"id"         json:"id"`
	Identifier string     `db:"identifier" json:"identifier"`
	Value      string     `db:"value"      json:"value"`
	ExpiresAt  time.Time  `db:"expires_at" json:"expiresAt"`
	CreatedAt  *time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt  *time.Time `db:"updated_at" json:"updatedAt"`
}

// CompanyMembership links a principal (user or agent) to a company.
type CompanyMembership struct {
	UUID           string    `db:"uuid"            json:"uuid"`
	CompanyUUID    string    `db:"company_uuid"    json:"companyUuid"`
	PrincipalType  string    `db:"principal_type"  json:"principalType"`
	PrincipalID    string    `db:"principal_id"    json:"principalId"`
	Status         string    `db:"status"          json:"status"`
	MembershipRole *string   `db:"membership_role" json:"membershipRole"`
	CreatedAt      time.Time `db:"created_at"      json:"createdAt"`
	UpdatedAt      time.Time `db:"updated_at"      json:"updatedAt"`
}

// InstanceUserRole stores instance-wide roles for users.
type InstanceUserRole struct {
	UUID      string    `db:"uuid"       json:"uuid"`
	UserID    string    `db:"user_id"    json:"userId"`
	Role      string    `db:"role"       json:"role"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

// Invite represents an invitation to join a company or instance.
type Invite struct {
	UUID              string          `db:"uuid"                json:"uuid"`
	CompanyUUID       *string         `db:"company_uuid"        json:"companyUuid"`
	InviteType        string          `db:"invite_type"         json:"inviteType"`
	TokenHash         string          `db:"token_hash"          json:"tokenHash"`
	AllowedJoinTypes  string          `db:"allowed_join_types"  json:"allowedJoinTypes"`
	DefaultsPayload   []byte          `db:"defaults_payload"    json:"defaultsPayload"`
	ExpiresAt         time.Time       `db:"expires_at"          json:"expiresAt"`
	InvitedByUserID   *string         `db:"invited_by_user_id"  json:"invitedByUserId"`
	RevokedAt         *time.Time      `db:"revoked_at"          json:"revokedAt"`
	AcceptedAt        *time.Time      `db:"accepted_at"         json:"acceptedAt"`
	CreatedAt         time.Time       `db:"created_at"          json:"createdAt"`
	UpdatedAt         time.Time       `db:"updated_at"          json:"updatedAt"`
}

// JoinRequest represents a request to join a company via an invite.
type JoinRequest struct {
	UUID                    string     `db:"uuid"                      json:"uuid"`
	InviteUUID              string     `db:"invite_uuid"               json:"inviteUuid"`
	CompanyUUID             string     `db:"company_uuid"              json:"companyUuid"`
	RequestType             string     `db:"request_type"              json:"requestType"`
	Status                  string     `db:"status"                    json:"status"`
	RequestIP               string     `db:"request_ip"                json:"requestIp"`
	RequestingUserID        *string    `db:"requesting_user_id"        json:"requestingUserId"`
	RequestEmailSnapshot    *string    `db:"request_email_snapshot"    json:"requestEmailSnapshot"`
	AgentName               *string    `db:"agent_name"                json:"agentName"`
	AdapterType             *string    `db:"adapter_type"              json:"adapterType"`
	Capabilities            *string    `db:"capabilities"              json:"capabilities"`
	AgentDefaultsPayload    []byte     `db:"agent_defaults_payload"    json:"agentDefaultsPayload"`
	CreatedAgentUUID        *string    `db:"created_agent_uuid"        json:"createdAgentUuid"`
	ApprovedByUserID        *string    `db:"approved_by_user_id"       json:"approvedByUserId"`
	ApprovedAt              *time.Time `db:"approved_at"               json:"approvedAt"`
	RejectedByUserID        *string    `db:"rejected_by_user_id"       json:"rejectedByUserId"`
	RejectedAt              *time.Time `db:"rejected_at"               json:"rejectedAt"`
	ClaimSecretHash         *string    `db:"claim_secret_hash"         json:"claimSecretHash"`
	ClaimSecretExpiresAt    *time.Time `db:"claim_secret_expires_at"   json:"claimSecretExpiresAt"`
	ClaimSecretConsumedAt   *time.Time `db:"claim_secret_consumed_at"  json:"claimSecretConsumedAt"`
	CreatedAt               time.Time  `db:"created_at"                json:"createdAt"`
	UpdatedAt               time.Time  `db:"updated_at"                json:"updatedAt"`
}

// AgentAPIKey represents an API key associated with an agent.
type AgentAPIKey struct {
	UUID        string     `db:"uuid"         json:"uuid"`
	AgentUUID   string     `db:"agent_uuid"   json:"agentUuid"`
	CompanyUUID string     `db:"company_uuid" json:"companyUuid"`
	Name        string     `db:"name"         json:"name"`
	KeyHash     string     `db:"key_hash"     json:"keyHash"`
	LastUsedAt  *time.Time `db:"last_used_at" json:"lastUsedAt"`
	RevokedAt   *time.Time `db:"revoked_at"   json:"revokedAt"`
	CreatedAt   time.Time  `db:"created_at"   json:"createdAt"`
}
