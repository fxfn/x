package main

import (
	"testing"

	"github.com/fxfn/x/inject"
)

func TestWithNested(t *testing.T) {
	container := inject.Default()
	inject.Register[SqliteDbProvider](container, NewSqliteDbProvider)
	inject.Register[UserRepository](container, NewUserRepository)

	userRepository := inject.Get[UserRepository](container)

	t.Run("should have a db", func(t *testing.T) {
		if userRepository.db == (SqliteDbProvider{}) {
			t.Errorf("userRepository.db is nil")
		}
	})

	t.Run("should have a db.id", func(t *testing.T) {
		if userRepository.db.id == "" {
			t.Errorf("userRepository.db.id is empty")
		}
	})

	t.Run("should have a db.connectionString", func(t *testing.T) {
		if userRepository.db.connectionString == "" {
			t.Errorf("userRepository.db.connectionString is empty")
		}
	})
}
