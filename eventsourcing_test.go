package eventsourcing_test

import (
	"eventsourcing"
	"fmt"
	"testing"
)

// Person aggregate
type Person struct {
	aggregateRoot eventsourcing.AggregateRoot
	name          string
	age           int
	dead          int
}

// Born event
type Born struct {
	name string
}

// AgedOneYear event
type AgedOneYear struct {
}

// CreatePerson constructor for the Person
func CreatePerson(name string) (*Person, error) {
	if name == "" {
		return nil, fmt.Errorf("name can't be blank")
	}

	person := &Person{}
	person.aggregateRoot.TrackChange(*person, Born{name: name}, person.transition)
	return person, nil
}

// CreatePersonWithID constructor for the Person that sets the aggregate id from the outside
func CreatePersonWithID(id, name string) (*Person, error) {
	if name == "" {
		return nil, fmt.Errorf("name can't be blank")
	}

	person := &Person{}
	err := person.aggregateRoot.SetID(id)
	if err == eventsourcing.ErrAggregateAlreadyExists {
		return nil, err
	} else if err != nil {
		panic(err)
	}
	person.aggregateRoot.TrackChange(*person, Born{name: name}, person.transition)
	return person, nil
}

// GrowOlder command
func (person *Person) GrowOlder() {
	person.aggregateRoot.TrackChange(*person, AgedOneYear{}, person.transition)

}

// transition the person state dependent on the events
func (person *Person) transition(event eventsourcing.Event) {
	switch e := event.Data.(type) {

	case Born:
		person.age = 0
		person.name = e.name

	case AgedOneYear:
		person.age += 1
	}
}

func TestCreateNewPerson(t *testing.T) {
	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal("Error when creating person", err.Error())
	}

	if person.name != "kalle" {
		t.Fatal("Wrong person name")
	}

	if person.age != 0 {
		t.Fatal("Wrong person age")
	}

	if len(person.aggregateRoot.Changes()) != 1 {
		t.Fatal("There should be one event on the person aggregateRoot")
	}

	if person.aggregateRoot.Version() != 1 {
		t.Fatal("Wrong version on the person aggregateRoot", person.aggregateRoot.Version())
	}
}

func TestCreateNewPersonWithIDFromOutside(t *testing.T) {
	id := "123"
	person, err := CreatePersonWithID(id, "kalle")
	if err != nil {
		t.Fatal("Error when creating person", err.Error())
	}

	if person.aggregateRoot.ID() != id {
		t.Fatal("Wrong aggregate id on the person aggregateRoot", person.aggregateRoot.ID())
	}
}

func TestBlankName(t *testing.T) {
	_, err := CreatePerson("")
	if err == nil {
		t.Fatal("The constructor should return error on blank name")
	}

}

func TestSetIDOnExistingPerson(t *testing.T) {
	person, err := CreatePerson("Kalle")
	if err != nil {
		t.Fatal("The constructor returned error")
	}

	err = person.aggregateRoot.SetID("new_id")
	if err == nil {
		t.Fatal("Should not be possible to set id on already existing person")
	}

}

func TestPersonAgedOneYear(t *testing.T) {
	person, _ := CreatePerson("kalle")
	person.GrowOlder()

	if len(person.aggregateRoot.Changes()) != 2 {
		t.Fatal("There should be two event on the person aggregateRoot", person.aggregateRoot.Changes())
	}

	if person.aggregateRoot.Changes()[len(person.aggregateRoot.Changes())-1].Reason != "AgedOneYear" {
		t.Fatal("The last event reason should be AgedOneYear", person.aggregateRoot.Changes()[len(person.aggregateRoot.Changes())-1].Reason)
	}
}

func TestPersonGrewTenYears(t *testing.T) {
	person, _ := CreatePerson("kalle")
	for i := 1; i <= 10; i++ {
		person.GrowOlder()
	}

	if person.age != 10 {
		t.Fatal("person has the wrong age")
	}
}