package jsonstore

import (
	"fmt"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/storage"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

// A RaidAlerts implements storage.RaidAlertsStore
type RaidAlerts struct {
	users storage.UsersStore
}

func (r RaidAlerts) get(steamID uint64) (types.RaidNotification, error) {
	ra, found := raidAlerts.d[steamID]
	if !found {
		err := raidAlerts.driver.Read(raidAlerts.collection, fmt.Sprintf("%d", steamID), &ra)
		if err != nil {
			return types.RaidNotification{}, err
		}
	}
	return ra, nil
}

// AddInfo implements storage.RaidAlertsStore.AddInfo
func (r RaidAlerts) AddInfo(alertIn time.Duration, ed types.EntityDeath) error {
	for _, steamID := range ed.Owners {
		var u types.User
		err := r.users.Get(steamID, &u)
		if err == nil {
			ra, err := r.get(u.SteamID)
			if err != nil {
				ra.GridPositions = []string{}
				ra.SteamID = u.SteamID
				ra.Items = map[string]int{}
				ra.AlertAt = time.Now().UTC().Add(alertIn)
			}

			items := ra.Items[ed.Name]
			ra.Items[ed.Name] = items + 1

			ra.GridPositions = append(ra.GridPositions, ed.GridPos)

			ra.GridPositions = func(gpos []string) []string {
				ps := make([]string, 0, len(gpos))
				m := make(map[string]bool)
				for _, val := range gpos {
					if _, ok := m[val]; !ok {
						m[val] = true
						ps = append(ps, val)
					}
				}
				return ps
			}(ra.GridPositions)
			err = raidAlerts.driver.Write(raidAlerts.collection, fmt.Sprintf("%d", u.SteamID), &ra)
			raidAlerts.d[u.SteamID] = ra
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetReady implements storage.RaidAlertsStore.GetReady
func (r RaidAlerts) GetReady(alerts *[]types.RaidNotification) error {
	for _, v := range raidAlerts.d {
		// fmt.Printf("%d, %s\n", time.Now().UTC().Sub(v.AlertAt).Seconds(), v.DiscordID)
		if int(time.Now().UTC().Sub(v.AlertAt).Seconds()) > 0 {
			*alerts = append(*alerts, v)
			err := r.Remove(v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Remove implements storage.RaidAlertsStore.Remove
func (r RaidAlerts) Remove(alert types.RaidNotification) error {
	err := raidAlerts.driver.Delete(raidAlerts.collection, fmt.Sprintf("%d", alert.SteamID))
	if err != nil {
		return err
	}
	delete(raidAlerts.d, alert.SteamID)
	return nil
}
