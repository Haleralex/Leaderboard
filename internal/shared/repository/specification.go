package repository

import (
	"gorm.io/gorm"
)

// Specification defines a business rule that can be checked against an entity
// It encapsulates query logic in a reusable and composable way
type Specification[T any] interface {
	// Apply applies the specification to a GORM query
	Apply(db *gorm.DB) *gorm.DB

	// IsSatisfiedBy checks if an entity satisfies this specification (optional, for in-memory filtering)
	IsSatisfiedBy(entity T) bool
}

// BaseSpecification provides default implementation
type BaseSpecification[T any] struct{}

func (s BaseSpecification[T]) IsSatisfiedBy(entity T) bool {
	// Default: always satisfied (override in concrete implementations)
	return true
}

// CompositeSpecification allows combining specifications
type CompositeSpecification[T any] struct {
	BaseSpecification[T]
}

// And combines two specifications with AND logic
type AndSpecification[T any] struct {
	CompositeSpecification[T]
	left  Specification[T]
	right Specification[T]
}

func NewAndSpecification[T any](left, right Specification[T]) Specification[T] {
	return &AndSpecification[T]{
		left:  left,
		right: right,
	}
}

func (s *AndSpecification[T]) Apply(db *gorm.DB) *gorm.DB {
	db = s.left.Apply(db)
	db = s.right.Apply(db)
	return db
}

func (s *AndSpecification[T]) IsSatisfiedBy(entity T) bool {
	return s.left.IsSatisfiedBy(entity) && s.right.IsSatisfiedBy(entity)
}

// Or combines two specifications with OR logic
type OrSpecification[T any] struct {
	CompositeSpecification[T]
	left  Specification[T]
	right Specification[T]
}

func NewOrSpecification[T any](left, right Specification[T]) Specification[T] {
	return &OrSpecification[T]{
		left:  left,
		right: right,
	}
}

func (s *OrSpecification[T]) Apply(db *gorm.DB) *gorm.DB {
	// OR requires subquery or raw SQL
	var leftQuery, rightQuery *gorm.DB
	leftQuery = s.left.Apply(db.Session(&gorm.Session{}))
	rightQuery = s.right.Apply(db.Session(&gorm.Session{}))

	// Combine with OR
	return db.Where(leftQuery).Or(rightQuery)
}

func (s *OrSpecification[T]) IsSatisfiedBy(entity T) bool {
	return s.left.IsSatisfiedBy(entity) || s.right.IsSatisfiedBy(entity)
}

// Not negates a specification
type NotSpecification[T any] struct {
	CompositeSpecification[T]
	spec Specification[T]
}

func NewNotSpecification[T any](spec Specification[T]) Specification[T] {
	return &NotSpecification[T]{
		spec: spec,
	}
}

func (s *NotSpecification[T]) Apply(db *gorm.DB) *gorm.DB {
	// NOT requires wrapping in NOT()
	subQuery := s.spec.Apply(db.Session(&gorm.Session{}))
	return db.Not(subQuery)
}

func (s *NotSpecification[T]) IsSatisfiedBy(entity T) bool {
	return !s.spec.IsSatisfiedBy(entity)
}

// Helper functions for fluent API
func And[T any](specs ...Specification[T]) Specification[T] {
	if len(specs) == 0 {
		return nil
	}
	if len(specs) == 1 {
		return specs[0]
	}

	result := specs[0]
	for i := 1; i < len(specs); i++ {
		result = NewAndSpecification(result, specs[i])
	}
	return result
}

func Or[T any](specs ...Specification[T]) Specification[T] {
	if len(specs) == 0 {
		return nil
	}
	if len(specs) == 1 {
		return specs[0]
	}

	result := specs[0]
	for i := 1; i < len(specs); i++ {
		result = NewOrSpecification(result, specs[i])
	}
	return result
}

func Not[T any](spec Specification[T]) Specification[T] {
	return NewNotSpecification(spec)
}
