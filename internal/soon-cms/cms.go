package retro

import (
	"context"
	"fmt"
	"strings"

	"github.com/bugfixes/go-bugfixes/logs"
	flagsService "github.com/flags-gg/go-flags"
	ConfigBuilder "github.com/keloran/go-config"
)

type CMS struct {
	Config       *ConfigBuilder.Config
	Flags        *flagsService.Client
	CTX          context.Context
	ErrorChannel chan error
}

type Service struct {
	ID          int            `json:"-"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	URL         string         `json:"url"`
	LaunchDate  LaunchDate     `json:"launchDate"`
	Progress    int            `json:"progress"`
	Progress2   float32        `json:"progress2"`
	Icon        string         `json:"icon"`
	FullDesc    string         `json:"fullDesc"`
	Uptime      string         `json:"uptime"`
	Launched    bool           `json:"launched"`
	Links       []ProjectLink  `json:"links,omitempty"`
	Roadmap     []RoadmapItem  `json:"roadmap,omitempty"`
}

type LaunchDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

type ProjectLink struct {
	LinkType string `json:"type"`
	URL      string `json:"url"`
	Label    string `json:"label,omitempty"`
}

type RoadmapItem struct {
	Name        string  `json:"name"`
	TargetDate  *string `json:"targetDate,omitempty"`
	ReleaseDate *string `json:"releaseDate,omitempty"`
	Completed   bool    `json:"completed"`
	SortOrder   int     `json:"sortOrder"`
}

func NewCMS(config *ConfigBuilder.Config, flags *flagsService.Client, errChan chan error) *CMS {
	return &CMS{
		Config:       config,
		Flags:        flags,
		CTX:          context.Background(),
		ErrorChannel: errChan,
	}
}

func (c CMS) getServices() ([]Service, error) {
	services := make([]Service, 0)

	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return services, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	rows, err := db.Query(c.CTX, "SELECT id, name, description, launch_year, launch_month, launch_day, url, progress, icon, full_description, uptime, live FROM services WHERE started = true")
	if err != nil {
		return nil, logs.Error(err)
	}
	defer rows.Close()

	for rows.Next() {
		var service Service
		if err := rows.Scan(
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
			&service.Uptime,
			&service.Launched); err != nil {
			return nil, logs.Error(err)
		}
		service.Uptime = fmt.Sprintf("https://uptime.chewedfeed.com/status/%s", service.Uptime)
		services = append(services, service)
	}

	for i, svc := range services {
		links, err := c.getServiceLinks(svc.ID)
		if err != nil {
			return services, logs.Error(err)
		}
		services[i].Links = links
	}

	return services, nil
}

func (c CMS) getService(name string) (Service, error) {
	var service Service

	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return service, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
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
		return Service{}, logs.Error(err)
	}

	service.Uptime = fmt.Sprintf("https://uptime.chewedfeed.com/status/%s", service.Uptime)

	prog, err := c.getServiceProgress(service.ID)
	if err != nil {
		return service, logs.Error(err)
	}
	service.Progress2 = prog

	links, err := c.getServiceLinks(service.ID)
	if err != nil {
		return service, logs.Error(err)
	}
	service.Links = links

	roadmap, err := c.getServiceRoadmap(service.ID)
	if err != nil {
		return service, logs.Error(err)
	}
	service.Roadmap = roadmap

	return service, nil
}

func (c CMS) getServiceProgress(id int) (float32, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return 0, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()
	rows, err := db.Query(c.CTX, "SELECT completed FROM launch_task WHERE service_id = $1", id)
	if err != nil {
		return 0, logs.Error(err)
	}
	totalRows := 0
	completedRows := 0

	defer rows.Close()
	for rows.Next() {
		var completed bool
		if err := rows.Scan(&completed); err != nil {
			return 0, logs.Error(err)
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

func (c CMS) getServiceLinks(serviceID int) ([]ProjectLink, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	rows, err := db.Query(c.CTX, "SELECT link_type, url, COALESCE(label, '') FROM project_links WHERE service_id = $1", serviceID)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer rows.Close()

	var links []ProjectLink
	for rows.Next() {
		var link ProjectLink
		if err := rows.Scan(&link.LinkType, &link.URL, &link.Label); err != nil {
			return nil, logs.Error(err)
		}
		links = append(links, link)
	}

	return links, nil
}

func (c CMS) getServiceRoadmap(serviceID int) ([]RoadmapItem, error) {
	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	rows, err := db.Query(c.CTX, "SELECT name, target_date::text, release_date::text, completed, sort_order FROM project_roadmap WHERE service_id = $1 ORDER BY sort_order", serviceID)
	if err != nil {
		return nil, logs.Error(err)
	}
	defer rows.Close()

	var roadmap []RoadmapItem
	for rows.Next() {
		var item RoadmapItem
		if err := rows.Scan(&item.Name, &item.TargetDate, &item.ReleaseDate, &item.Completed, &item.SortOrder); err != nil {
			return nil, logs.Error(err)
		}
		roadmap = append(roadmap, item)
	}

	return roadmap, nil
}

func (c CMS) AllowedOrigins() ([]string, error) {
	origins := make([]string, 0)

	db, err := c.Config.Database.GetPGXClient(c.CTX)
	if err != nil {
		return origins, logs.Error(err)
	}
	defer func() {
		if err := db.Close(c.CTX); err != nil {
			c.ErrorChannel <- logs.Error(err)
		}
	}()

	rows, err := db.Query(c.CTX, "SELECT non_url, alternatives FROM services")
	if err != nil {
		return nil, logs.Error(err)
	}
	type service struct {
		URL  *string
		Alts *string
	}

	defer rows.Close()
	for rows.Next() {
		var s service
		if err := rows.Scan(&s.URL, &s.Alts); err != nil {
			return nil, logs.Error(err)
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
