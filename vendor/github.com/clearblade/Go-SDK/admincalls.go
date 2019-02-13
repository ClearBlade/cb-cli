package GoSDK

import (
	"fmt"
)

//
// DevClient must already be platform admin to use these endpoints
//

func (d *DevClient) PromoteDevToPlatformAdmin(email string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"email": email,
	}

	resp, err := post(d, d.preamble()+"/promotedev", data, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error promoting %s to admin: %v", email, resp.Body)
	}
	return nil
}

func (d *DevClient) DemoteDevFromPlatformAdmin(email string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"email": email,
	}

	resp, err := post(d, d.preamble()+"/demotedev", data, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error demoting %s to admin: %v", email, resp.Body)
	}
	return nil
}

func (d *DevClient) ResetDevelopersPassword(email, newPassword string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"email":        email,
		"new_password": newPassword,
	}

	resp, err := post(d, d.preamble()+"/resetpassword", data, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error resetting %s's password: %v", email, resp.Body)
	}
	return nil
}

func (d *DevClient) GetSystemAnalytics(systemKey string) (interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	analyticsEndpoint := d.preamble() + "/platform/system/" + systemKey
	resp, err := get(d, analyticsEndpoint, nil, creds, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting %s's analytics: %v", systemKey, resp.Body)
	}
	return resp.Body, nil
}

func (d *DevClient) GetAllSystemsAnalytics(query string) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	analyticsEndpoint := d.preamble() + "/platform/systems"
	q := make(map[string]string)
	if len(query) != 0 {
		q["query"] = query
	} else {
		q = nil
	}
	resp, err := get(d, analyticsEndpoint, q, creds, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting analytics: %v", resp.Body)
	}
	return resp.Body.([]interface{}), nil
}

func (d *DevClient) DisableSystem(systemKey string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	disableSystemEndpoint := d.preamble() + "/platform/" + systemKey
	resp, err := delete(d, disableSystemEndpoint, nil, creds, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error disabling system %s: %v", systemKey, resp.Body)
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) EnableSystem(systemKey string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	disableSystemEndpoint := d.preamble() + "/platform/" + systemKey
	resp, err := post(d, disableSystemEndpoint, nil, creds, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error disabling system %s: %v", systemKey, resp.Body)
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) GetDeveloper(devEmail string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	q := make(map[string]string)
	q["developer"] = devEmail

	developerEndpoint := d.preamble() + "/platform/developer"
	resp, err := get(d, developerEndpoint, q, creds, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting developer %s: %v", devEmail, resp.Body)
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) GetAllDevelopers() (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	allDevelopersEndpoint := d.preamble() + "/platform/developers"
	resp, err := get(d, allDevelopersEndpoint, nil, creds, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting all developers: %v", resp.Body)
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) SetDeveloper(email string, admin, disabled bool) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{
		"email":    email,
		"admin":    admin,
		"disabled": disabled,
	}

	developerEndpoint := d.preamble() + "/platform/developer"
	resp, err := post(d, developerEndpoint, data, creds, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error setting developer %s: %v", email, resp.Body)
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) GetMetrics(metricType string) (interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	metricsEndpoint := d.preamble() + "/platform/" + metricType
	resp, err := get(d, metricsEndpoint, nil, creds, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting metric %s: %v", metricType, resp.Body)
	}
	return resp.Body, nil
}
