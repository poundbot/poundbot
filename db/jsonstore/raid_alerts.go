package jsonstore

import (
	"time"

	"bitbucket.org/mrpoundsign/poundbot/db"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

// A RaidAlerts implements db.RaidAlertsStore
type RaidAlerts struct {
	users db.UsersStore
}

func (r RaidAlerts) get(discordID string) (types.RaidNotification, error) {
	ra, found := raidAlerts.d[discordID]
	if !found {
		err := raidAlerts.driver.Read(raidAlerts.collection, discordID, &ra)
		if err != nil {
			return types.RaidNotification{}, err
		}
	}
	return ra, nil
}

// AddInfo implements db.RaidAlertsStore.AddInfo
func (r RaidAlerts) AddInfo(ed types.EntityDeath) error {
	for _, steamID := range ed.Owners {
		u, err := r.users.Get(types.SteamInfo{SteamID: steamID})
		if err == nil {
			ra, err := r.get(u.DiscordID)
			if err != nil {
				ra.GridPositions = []string{}
				ra.DiscordID = u.DiscordID
				ra.Items = map[string]int{}
				ra.AlertAt = time.Now().UTC().Add(10 * time.Second)
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
			err = raidAlerts.driver.Write(raidAlerts.collection, u.DiscordID, &ra)
			raidAlerts.d[u.DiscordID] = ra
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetReady implements db.RaidAlertsStore.GetReady
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

// Remove implements db.RaidAlertsStore.Remove
func (r RaidAlerts) Remove(alert types.RaidNotification) error {
	err := raidAlerts.driver.Delete(raidAlerts.collection, alert.DiscordID)
	if err != nil {
		return err
	}
	delete(raidAlerts.d, alert.DiscordID)
	return nil
}
