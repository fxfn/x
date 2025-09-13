package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fxfn/x/inject"
)

type SqliteDbProvider struct {
	id               string
	connectionString string
}

type Repository[T any] struct {
	db SqliteDbProvider
}

type User struct {
	ID   string
	Name string
}

type UserRepository struct {
	Repository[User]
}

func generateRandomId() string {
	// use no third party libraries to generate a random alphanumeric id
	source := rand.NewSource(time.Now().UnixNano())
	return fmt.Sprintf("%d", rand.New(source).Intn(1000000))
}

func NewSqliteDbProvider(c *inject.Container) SqliteDbProvider {
	return SqliteDbProvider{
		id:               generateRandomId(),
		connectionString: "file://./sqlite.db",
	}
}

func NewUserRepository(c *inject.Container) UserRepository {
	return UserRepository{
		Repository: Repository[User]{
			db: inject.Get[SqliteDbProvider](c),
		},
	}
}

type Customer struct {
	ID   string
	Name string
}

type CustomerRepository struct {
	Repository[Customer]
}

func NewCustomerRepository(c *inject.Container) CustomerRepository {
	return CustomerRepository{
		Repository: Repository[Customer]{
			db: inject.Get[SqliteDbProvider](c),
		},
	}
}

func main() {
	container := inject.NewContainer()

	inject.RegisterSingleton[SqliteDbProvider](container, NewSqliteDbProvider)
	inject.Register[UserRepository](container, NewUserRepository)
	inject.Register[CustomerRepository](container, NewCustomerRepository)

	userRepository := inject.Get[UserRepository](container)
	customerRepository := inject.Get[CustomerRepository](container)

	fmt.Printf("%+v\n", userRepository)
	fmt.Printf("%+v\n", customerRepository)
}
