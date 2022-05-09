package zenefits

import (
	"github.com/pomerium/datasource/internal/util"
)

// Response is generic response from zenefits API
type Response[T Kind] struct {
	Error  string  `json:"error"`
	Status int     `json:"status"`
	Data   List[T] `json:"data"`
}

// List represents pageable response
type List[T Kind] struct {
	Items   []T       `json:"data"`
	NextURL *util.URL `json:"next_url"`
}

// Kind represents an object kind
type Kind interface {
	Person | Department | Location | Vacation
}

// Person see https://developers.zenefits.com/docs/people
type Person struct {
	ID string `json:"id" mapstructure:"zenefits_id"`

	FirstName     string `json:"first_name" mapstructure:"first_name"`
	MiddleName    string `json:"middle_name" mapstructure:"middle_name,omitempty"`
	Lastname      string `json:"last_name" mapstructure:"last_name"`
	PreferredName string `json:"preferred_name" mapstructure:"preferred_name,omitempty"`

	WorkEmail     string `json:"work_email" mapstructure:"id,omitempty"`
	WorkPhone     string `json:"work_phone" mapstructure:"work_phone,omitempty"`
	PersonalEmail string `json:"personal_email" mapstructure:"personal_email,omitempty"`
	PersonalPhone string `json:"personal_phone" mapstructure:"personal_phone,omitempty"`

	Department *Department `json:"department" mapstructure:",squash,omitempty"`
	Title      string      `json:"title" mapstructure:"title"`
	Status     string      `json:"status" mapstructure:"status"`
	Type       string      `json:"type" mapstructure:"type"`

	Location *Location `json:"location" mapstructure:",squash,omitempty"`

	Country string `json:"country" mapstructure:"home_country,omitempty"`
	State   string `json:"state" mapstructure:"home_state,omitempty"`
	City    string `json:"city" mapstructure:"home_city,omitempty"`
	Street1 string `json:"street1" mapstructure:"home_street1,omitempty"`
	Street2 string `json:"street2" mapstructure:"home_street2,omitempty"`
}

// Vacation see https://developers.zenefits.com/docs/vacation-requests
type Vacation struct {
	Person Person `json:"person"`
}

// Department see https://developers.zenefits.com/docs/department
type Department struct {
	ID   string `json:"id" mapstructure:"dept_id"`
	Name string `json:"name" mapstructure:"dept_name"`
}

// Location see https://developers.zenefits.com/docs/location
type Location struct {
	ID      string `json:"id" mapstructure:"loc_id"`
	Name    string `json:"name" mapstructure:"loc_name"`
	Country string `json:"country" mapstructure:"loc_country"`
	City    string `json:"city" mapstructure:"loc_city"`
	State   string `json:"state" mapstructure:"loc_state"`
	Street1 string `json:"street1" mapstructure:"loc_street1"`
	Street2 string `json:"street2" mapstructure:"loc_street2"`
	Phone   string `json:"phone" mapstructure:"loc_phone"`
}
