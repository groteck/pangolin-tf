package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type apiResponse struct {
	Data    json.RawMessage `json:"data"`
	Success bool            `json:"success"`
	Message string          `json:"message"`
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("API failure: %s", apiResp.Message)
	}

	return apiResp.Data, nil
}

// Role definitions
type Role struct {
	ID          int    `json:"roleId,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *Client) CreateRole(orgID string, role *Role) (*Role, error) {
	path := fmt.Sprintf("/org/%s/role", orgID)
	body := map[string]interface{}{
		"name":        role.Name,
		"description": role.Description,
	}
	data, err := c.doRequest("PUT", path, body)
	if err != nil {
		return nil, err
	}
	var out Role
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) GetRole(orgID string, roleID int) (*Role, error) {
	path := fmt.Sprintf("/role/%d", roleID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var out Role
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) UpdateRole(orgID string, roleID int, role *Role) (*Role, error) {
	path := fmt.Sprintf("/role/%d", roleID)
	body := map[string]interface{}{
		"name":        role.Name,
		"description": role.Description,
	}
	data, err := c.doRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	var out Role
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) DeleteRole(orgID string, roleID int) error {
	path := fmt.Sprintf("/role/%d", roleID)
	// Workaround: Pangolin requires a replacement role ID for users in the deleted role.
	// We use ID 2 (Member) which is standard in a fresh org.
	body := map[string]interface{}{
		"roleId": "2",
	}
	_, err := c.doRequest("DELETE", path, body)
	return err
}

func (c *Client) ListRoles(orgID string) ([]Role, error) {
	path := fmt.Sprintf("/org/%s/roles", orgID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Roles []Role `json:"roles"`
	}
	err = json.Unmarshal(data, &wrapper)
	return wrapper.Roles, err
}

// Site definitions
type Site struct {
	ID      int    `json:"siteId"`
	Name    string `json:"name"`
	NewtID  string `json:"newtId,omitempty"`
	Secret  string `json:"secret,omitempty"`
	Address string `json:"address,omitempty"`
	Type    string `json:"type,omitempty"`
}

func (c *Client) ListSites(orgID string) ([]Site, error) {
	path := fmt.Sprintf("/org/%s/sites", orgID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Sites []Site `json:"sites"`
	}
	err = json.Unmarshal(data, &wrapper)
	return wrapper.Sites, err
}

func (c *Client) CreateSite(orgID string, site *Site) (*Site, error) {
	path := fmt.Sprintf("/org/%s/site", orgID)
	body := map[string]interface{}{
		"name":   site.Name,
		"newtId": site.NewtID,
		"secret": site.Secret,
		"type":   site.Type,
	}
	if site.Address != "" {
		body["address"] = site.Address
	}
	data, err := c.doRequest("PUT", path, body)
	if err != nil {
		return nil, err
	}
	var out Site
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) GetSite(siteID int) (*Site, error) {
	path := fmt.Sprintf("/site/%d", siteID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var out Site
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) UpdateSite(siteID int, site *Site) (*Site, error) {
	path := fmt.Sprintf("/site/%d", siteID)
	body := map[string]interface{}{
		"name": site.Name,
	}
	if site.Address != "" {
		body["address"] = site.Address
	}
	data, err := c.doRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	var out Site
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) DeleteSite(siteID int) error {
	path := fmt.Sprintf("/site/%d", siteID)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

// SiteResource definitions
type SiteResource struct {
	ID                 int      `json:"siteResourceId,omitempty"`
	NiceID             string   `json:"niceId,omitempty"`
	Name               string   `json:"name"`
	Mode               string   `json:"mode"`
	SiteID             int      `json:"siteId"`
	Destination        string   `json:"destination"`
	Enabled            bool     `json:"enabled"`
	Alias              *string  `json:"alias,omitempty"`
	UserIDs            []string `json:"userIds"`
	RoleIDs            []int    `json:"roleIds"`
	ClientIDs          []int    `json:"clientIds"`
	TCPPortRangeString string   `json:"tcpPortRangeString,omitempty"`
	UDPPortRangeString string   `json:"udpPortRangeString,omitempty"`
	DisableIcmp        bool     `json:"disableIcmp,omitempty"`
}

func (c *Client) CreateSiteResource(orgID string, res *SiteResource) (*SiteResource, error) {
	path := fmt.Sprintf("/org/%s/site-resource", orgID)
	body := map[string]interface{}{
		"name":        res.Name,
		"mode":        res.Mode,
		"siteId":      res.SiteID,
		"destination": res.Destination,
		"enabled":     res.Enabled,
		"userIds":     res.UserIDs,
		"roleIds":     res.RoleIDs,
		"clientIds":   res.ClientIDs,
	}
	if res.Alias != nil {
		body["alias"] = *res.Alias
	}
	data, err := c.doRequest("PUT", path, body)
	if err != nil {
		return nil, err
	}
	var out SiteResource
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) GetSiteResource(orgID string, siteID int, resID int) (*SiteResource, error) {
	path := fmt.Sprintf("/site-resource/%d", resID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var out SiteResource
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) UpdateSiteResource(resID int, res *SiteResource) (*SiteResource, error) {
	path := fmt.Sprintf("/site-resource/%d", resID)
	body := map[string]interface{}{
		"name":        res.Name,
		"siteId":      res.SiteID,
		"mode":        res.Mode,
		"destination": res.Destination,
		"enabled":     res.Enabled,
		"userIds":     res.UserIDs,
		"roleIds":     res.RoleIDs,
		"clientIds":   res.ClientIDs,
	}
	if res.Alias != nil {
		body["alias"] = *res.Alias
	}
	data, err := c.doRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	var out SiteResource
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) DeleteSiteResource(resID int) error {
	path := fmt.Sprintf("/site-resource/%d", resID)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

func (c *Client) GetSiteResourceRoles(resID int) ([]int, error) {
	path := fmt.Sprintf("/site-resource/%d/roles", resID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Roles []struct {
			RoleID int `json:"roleId"`
		} `json:"roles"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(wrapper.Roles))
	seen := make(map[int]bool)
	for _, r := range wrapper.Roles {
		if !seen[r.RoleID] {
			ids = append(ids, r.RoleID)
			seen[r.RoleID] = true
		}
	}
	return ids, nil
}

func (c *Client) GetSiteResourceUsers(resID int) ([]string, error) {
	path := fmt.Sprintf("/site-resource/%d/users", resID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Users []struct {
			UserID string `json:"userId"`
		} `json:"users"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	ids := make([]string, len(wrapper.Users))
	for i, u := range wrapper.Users {
		ids[i] = u.UserID
	}
	return ids, nil
}

func (c *Client) GetSiteResourceClients(resID int) ([]int, error) {
	path := fmt.Sprintf("/site-resource/%d/clients", resID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Clients []struct {
			ClientID int `json:"clientId"`
		} `json:"clients"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	ids := make([]int, len(wrapper.Clients))
	for i, cl := range wrapper.Clients {
		ids[i] = cl.ClientID
	}
	return ids, nil
}

// Resource definitions
type Resource struct {
	ID        int    `json:"resourceId,omitempty"`
	Name      string `json:"name"`
	Protocol  string `json:"protocol"`
	Http      bool   `json:"http"`
	Subdomain string `json:"subdomain"`
	DomainID  string `json:"domainId"`
}

func (c *Client) CreateResource(orgID string, res *Resource) (*Resource, error) {
	path := fmt.Sprintf("/org/%s/resource", orgID)
	data, err := c.doRequest("PUT", path, res)
	if err != nil {
		return nil, err
	}
	var out Resource
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) GetResource(resID int) (*Resource, error) {
	path := fmt.Sprintf("/resource/%d", resID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var out Resource
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) UpdateResource(resID int, res *Resource) (*Resource, error) {
	path := fmt.Sprintf("/resource/%d", resID)
	data, err := c.doRequest("POST", path, res)
	if err != nil {
		return nil, err
	}
	var out Resource
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) DeleteResource(resID int) error {
	path := fmt.Sprintf("/resource/%d", resID)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

// Target definitions
type Target struct {
	ID                  int            `json:"targetId,omitempty"`
	SiteID              int            `json:"siteId"`
	IP                  string         `json:"ip"`
	Port                int            `json:"port"`
	Method              *string        `json:"method,omitempty"`
	Enabled             bool           `json:"enabled"`
	HCEnabled           *bool          `json:"hcEnabled,omitempty"`
	HCPath              *string        `json:"hcPath,omitempty"`
	HCScheme            *string        `json:"hcScheme,omitempty"`
	HCMode              *string        `json:"hcMode,omitempty"`
	HCHostname          *string        `json:"hcHostname,omitempty"`
	HCPort              *int           `json:"hcPort,omitempty"`
	HCInterval          *int           `json:"hcInterval,omitempty"`
	HCUnhealthyInterval *int           `json:"hcUnhealthyInterval,omitempty"`
	HCTimeout           *int           `json:"hcTimeout,omitempty"`
	HCHeaders           []TargetHeader `json:"hcHeaders,omitempty"`
	HCFollowRedirects   *bool          `json:"hcFollowRedirects,omitempty"`
	HCMethod            *string        `json:"hcMethod,omitempty"`
	HCStatus            *int           `json:"hcStatus,omitempty"`
	HCTlsServerName     *string        `json:"hcTlsServerName,omitempty"`
	Path                *string        `json:"path,omitempty"`
	PathMatchType       *string        `json:"pathMatchType,omitempty"`
	RewritePath         *string        `json:"rewritePath,omitempty"`
	RewritePathType     *string        `json:"rewritePathType,omitempty"`
	Priority            *int           `json:"priority,omitempty"`
}

type TargetHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (c *Client) CreateTarget(resID int, target *Target) (*Target, error) {
	path := fmt.Sprintf("/resource/%d/target", resID)
	body := map[string]interface{}{
		"siteId":  target.SiteID,
		"ip":      target.IP,
		"port":    target.Port,
		"enabled": target.Enabled,
	}
	// Add other optional fields if needed...
	data, err := c.doRequest("PUT", path, body)
	if err != nil {
		return nil, err
	}
	var out Target
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) GetTarget(targetID int) (*Target, error) {
	path := fmt.Sprintf("/target/%d", targetID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var out Target
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) UpdateTarget(targetID int, target *Target) (*Target, error) {
	path := fmt.Sprintf("/target/%d", targetID)
	body := map[string]interface{}{
		"siteId":  target.SiteID,
		"ip":      target.IP,
		"port":    target.Port,
		"enabled": target.Enabled,
	}
	data, err := c.doRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	var out Target
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) DeleteTarget(targetID int) error {
	path := fmt.Sprintf("/target/%d", targetID)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}
