package retro

import (
	"context"
	"fmt"
	ConfigBuilder "github.com/keloran/go-config"
	"strings"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type CMS struct {
	Config       *ConfigBuilder.Config
	CTX          context.Context
	ErrorChannel chan error
}

type Service struct {
	ID          int
	Name        string     `json:"name"`
	Description string     `json:"description"`
	URL         string     `json:"url"`
	LaunchDate  LaunchDate `json:"launchDate"`
	Progress    int        `json:"progress"`
	Progress2   float32    `json:"progress2"`
	Icon        string     `json:"icon"`
	FullDesc    string     `json:"fullDesc"`
	Uptime      string     `json:"uptime"`
	Launched    bool       `json:"launched"`
}
type LaunchDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

func NewCMS(config *ConfigBuilder.Config, errChan chan error) *CMS {
	return &CMS{
		Config:       config,
		CTX:          context.Background(),
		ErrorChannel: errChan,
	}
}

func (c CMS) getServices() ([]Service, error) {
	services := make([]Service, 0)

	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return services, bugLog.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- bugLog.Error(err)
		}
	}()

	rows, err := db.Query(c.CTX, "SELECT name, description, launch_year, launch_month, launch_day, url, progress, icon, full_description, uptime, live FROM services WHERE started = true")
	if err != nil {
		return nil, bugLog.Error(err)
	}
	defer rows.Close()

	for rows.Next() {
		var service Service
		if err := rows.Scan(
			&service.Name,
			&service.Description,
			&service.LaunchDate.Year,
			&service.LaunchDate.Month,
			&service.LaunchDate.Day,
			&service.URL,
			&service.Progress,
			&service.Icon,
			&service.FullDesc,
			&service.Uptime,
			&service.Launched); err != nil {
			return nil, bugLog.Error(err)
		}
		service.Uptime = fmt.Sprintf("https://uptime.chewedfeed.com/status/%s", service.Uptime)
		services = append(services, service)
	}

	return services, nil
}

func (c CMS) getService(name string) (Service, error) {
	var service Service

	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return service, bugLog.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- bugLog.Error(err)
		}
	}()

	if err := db.QueryRow(c.CTX, "SELECT id, name, description, launch_year, launch_month, launch_day, url, progress, icon, full_description, uptime FROM services WHERE search_name = $1", name).Scan(
		&service.ID,
		&service.Name,
		&service.Description,
		&service.LaunchDate.Year,
		&service.LaunchDate.Month,
		&service.LaunchDate.Day,
		&service.URL,
		&service.Progress,
		&service.Icon,
		&service.FullDesc,
		&service.Uptime); err != nil {
		return Service{}, bugLog.Error(err)
	}

	service.Uptime = fmt.Sprintf("https://uptime.chewedfeed.com/status/%s", service.Uptime)

	prog, err := c.getServiceProgress(service.ID)
	if err != nil {
		return service, bugLog.Error(err)
	}
	service.Progress2 = prog

	return service, nil
}

func (c CMS) getServiceProgress(id int) (float32, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return 0, bugLog.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- bugLog.Error(err)
		}
	}()
	rows, err := db.Query(c.CTX, "SELECT completed FROM launch_task WHERE service_id = $1", id)
	if err != nil {
		return 0, bugLog.Error(err)
	}
	totalRows := 0
	completedRows := 0

	defer rows.Close()
	for rows.Next() {
		var completed bool
		if err := rows.Scan(&completed); err != nil {
			return 0, bugLog.Error(err)
		}
		totalRows++
		if completed {
			completedRows++
		}
	}
	if totalRows == 0 {
		return 0, nil
	}

	return (float32(completedRows) / float32(totalRows)) * 100, nil
}

func (c CMS) AllowedOrigins() ([]string, error) {
	origins := make([]string, 0)

	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return origins, bugLog.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- bugLog.Error(err)
		}
	}()

	rows, err := db.Query(c.CTX, "SELECT non_url, alternatives FROM services")
	if err != nil {
		return nil, bugLog.Error(err)
	}
	type service struct {
		URL  *string
		Alts *string
	}

	defer rows.Close()
	for rows.Next() {
		var s service
		if err := rows.Scan(&s.URL, &s.Alts); err != nil {
			return nil, bugLog.Error(err)
		}
		if s.Alts != nil {
			alts := strings.Split(*s.Alts, ",")
			for _, alt := range alts {
				if alt != "" {
					origins = append(origins, fmt.Sprintf("https://%s", alt), fmt.Sprintf("https://www.%s", alt))
				}
			}
		}
		origins = append(origins, fmt.Sprintf("https://%s", *s.URL), fmt.Sprintf("https://www.%s", *s.URL))
	}

	return origins, nil
}
