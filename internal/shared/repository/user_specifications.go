package repository

import (
	"strings"

	authmodels "leaderboard-service/internal/auth/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User Specifications

// UserByIDSpec filters users by ID
type UserByIDSpec struct {
	BaseSpecification[authmodels.User]
	ID uuid.UUID
}

func NewUserByIDSpec(id uuid.UUID) Specification[authmodels.User] {
	return &UserByIDSpec{ID: id}
}

func (s *UserByIDSpec) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("id = ?", s.ID)
}

func (s *UserByIDSpec) IsSatisfiedBy(user authmodels.User) bool {
	return user.ID == s.ID
}

// UserByEmailSpec filters users by email
type UserByEmailSpec struct {
	BaseSpecification[authmodels.User]
	Email string
}

func NewUserByEmailSpec(email string) Specification[authmodels.User] {
	return &UserByEmailSpec{Email: email}
}

func (s *UserByEmailSpec) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("email = ?", s.Email)
}

func (s *UserByEmailSpec) IsSatisfiedBy(user authmodels.User) bool {
	return user.Email == s.Email
}

// UserByNameSpec filters users by name (case-insensitive partial match)
type UserByNameSpec struct {
	BaseSpecification[authmodels.User]
	Name string
}

func NewUserByNameSpec(name string) Specification[authmodels.User] {
	return &UserByNameSpec{Name: name}
}

func (s *UserByNameSpec) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(s.Name)+"%")
}

func (s *UserByNameSpec) IsSatisfiedBy(user authmodels.User) bool {
	return strings.Contains(strings.ToLower(user.Name), strings.ToLower(s.Name))
}

// UserByEmailDomainSpec filters users by email domain
type UserByEmailDomainSpec struct {
	BaseSpecification[authmodels.User]
	Domain string
}

func NewUserByEmailDomainSpec(domain string) Specification[authmodels.User] {
	return &UserByEmailDomainSpec{Domain: domain}
}

func (s *UserByEmailDomainSpec) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("email LIKE ?", "%@"+s.Domain)
}

func (s *UserByEmailDomainSpec) IsSatisfiedBy(user authmodels.User) bool {
	return strings.HasSuffix(user.Email, "@"+s.Domain)
}

// UserCreatedAfterSpec filters users created after a specific time
type UserCreatedAfterSpec struct {
	BaseSpecification[authmodels.User]
	After string // ISO timestamp
}

func NewUserCreatedAfterSpec(after string) Specification[authmodels.User] {
	return &UserCreatedAfterSpec{After: after}
}

func (s *UserCreatedAfterSpec) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("created_at > ?", s.After)
}

func (s *UserCreatedAfterSpec) IsSatisfiedBy(user authmodels.User) bool {
	return user.CreatedAt.String() > s.After
}

// UserLimitSpec limits the number of results
type UserLimitSpec struct {
	BaseSpecification[authmodels.User]
	Limit int
}

func NewUserLimitSpec(limit int) Specification[authmodels.User] {
	return &UserLimitSpec{Limit: limit}
}

func (s *UserLimitSpec) Apply(db *gorm.DB) *gorm.DB {
	return db.Limit(s.Limit)
}

func (s *UserLimitSpec) IsSatisfiedBy(user authmodels.User) bool {
	return true // Limit doesn't affect individual entities
}

// UserOrderBySpec orders results
type UserOrderBySpec struct {
	BaseSpecification[authmodels.User]
	Field string
	Desc  bool
}

func NewUserOrderBySpec(field string, desc bool) Specification[authmodels.User] {
	return &UserOrderBySpec{Field: field, Desc: desc}
}

func (s *UserOrderBySpec) Apply(db *gorm.DB) *gorm.DB {
	order := s.Field
	if s.Desc {
		order += " DESC"
	}
	return db.Order(order)
}

func (s *UserOrderBySpec) IsSatisfiedBy(user authmodels.User) bool {
	return true // Order doesn't affect individual entities
}

// Convenience builders for common queries

// ActiveUsersSpec - example of composed specification
func ActiveUsersSpec(domain string) Specification[authmodels.User] {
	// Users from specific domain, ordered by creation date
	return And(
		NewUserByEmailDomainSpec(domain),
		NewUserOrderBySpec("created_at", true),
	)
}

// RecentUsersSpec - users created in last period
func RecentUsersSpec(after string, limit int) Specification[authmodels.User] {
	return And(
		NewUserCreatedAfterSpec(after),
		NewUserOrderBySpec("created_at", true),
		NewUserLimitSpec(limit),
	)
}

// SearchUsersSpec - search by name or email
func SearchUsersSpec(query string) Specification[authmodels.User] {
	return Or(
		NewUserByNameSpec(query),
		NewUserByEmailSpec(query),
	)
}
