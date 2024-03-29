package command

import (
	"RoleKeeper/cons"
	"RoleKeeper/cwlog"
	"RoleKeeper/db"
	"RoleKeeper/disc"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func handleComponet(s *discordgo.Session, i *discordgo.InteractionCreate) {

	data := i.MessageComponentData()
	guild := db.GuildLookupReadString(i.GuildID)
	if guild == nil {
		disc.EphemeralResponse(s, i, disc.DiscRed, "ERROR:", "This Discord guild is not in our database.", true)
		return
	}

	dirty := false
	errfound := false
	changes := ""

	if strings.EqualFold(data.CustomID, "assign-roles") {

		disc.EphemeralResponse(s, i, disc.DiscPurple, "Status:", "Loading role data...", false)

		for _, c := range data.Values {
			roleid, err := db.SnowflakeToInt(c)

			if err != nil {
				changes = changes + "Selection contained invalid role data!\n"
				errfound = true
				continue
			}

			found := -1
			for rpos, role := range guild.Roles {
				if role.ID == roleid {
					found = rpos
					break
				}
			}

			db.LookupRoleNames(s, guild)
			guild.Lock.RLock()
			roleName := db.IntToSnowflake(roleid)
			for _, role := range guild.Roles {
				if role.ID == roleid {
					roleName = role.Name
				}
			}
			guild.Lock.RUnlock()

			if found != -1 {

				if disc.UserHasRole(i, c) {
					err := disc.SmartRoleDelete(s, i.GuildID, i.Member.User.ID, c)
					if err != nil {
						changes = changes + "Unable to remove role: " + roleName + "\n"
						errfound = true
						continue
					} else {
						changes = changes + "Role removed: " + roleName + "\n"
						continue
					}
				} else {
					err := disc.SmartRoleAdd(s, i.GuildID, i.Member.User.ID, c)
					if err != nil {
						changes = changes + "Unable to assign role: " + roleName + "\n"
						errfound = true
						continue
					} else {
						changes = changes + "Role assigned: " + roleName + "\n"
						continue

					}
				}
			} else {
				changes = changes + "Role invalid: " + roleName + "\n"
				continue
			}

		}
	} else if strings.EqualFold(data.CustomID, "config-roles") {

		/* Check moderator privleges */
		roles := disc.GetGuildRoles(s, i.GuildID)
		for _, role := range roles {
			for _, mrole := range i.Member.Roles {
				if mrole == role.Name {
					if role.Permissions&(discordgo.PermissionAdministrator|
						discordgo.PermissionBanMembers|
						discordgo.PermissionManageRoles|
						discordgo.PermissionModerateMembers|
						discordgo.PermissionManageWebhooks|
						discordgo.PermissionManageServer) == 0 {
						return /* This user is not a moderator, don't even reply */
					}
				}
			}
		}

		disc.EphemeralResponse(s, i, disc.DiscPurple, "Status:", "Loading role data...", false)

		for _, c := range data.Values {

			roleid, err := db.SnowflakeToInt(c)

			if err != nil {
				changes = changes + "Selection contained invalid role data!\n"
				errfound = true
				continue
			}

			found := -1
			for rpos, role := range guild.Roles {
				if role.ID == roleid {
					found = rpos
					break
				}
			}

			/*
			 * If role exists, remove it
			 */
			if found != -1 {

				/* Lookup name before remove */
				db.LookupRoleNames(s, guild)
				guild.Lock.RLock()
				roleName := db.IntToSnowflake(roleid)
				for _, role := range guild.Roles {
					if role.ID == roleid {
						roleName = role.Name
					}
				}
				guild.Lock.RUnlock()
				time.Sleep(cons.LockRest)

				guild.Lock.Lock()
				guild.Roles = append(guild.Roles[:found], guild.Roles[found+1:]...)
				guild.Lock.Unlock()

				changes = changes + "Role removed: " + roleName + "\n"
				dirty = true //Save DB
				continue
			} else {

				/* Add role, if list isn't full */
				numRoles := len(guild.Roles)
				if numRoles < cons.LimitRoles {

					newRole := db.RoleData{ID: roleid}
					guild.Lock.Lock()
					guild.Roles = append(guild.Roles, newRole)
					guild.Modified = db.NowToCompact()
					guild.Lock.Unlock()

					time.Sleep(cons.LockRest)

					/* Lookup name after add */
					db.LookupRoleNames(s, guild)
					guild.Lock.RLock()
					roleName := db.IntToSnowflake(roleid)
					for _, role := range guild.Roles {
						if role.ID == roleid {
							roleName = role.Name
						}
					}
					guild.Lock.RUnlock()

					dirty = true //Save DB
					changes = changes + "Role added: " + roleName + "\n"
					continue

				} else {
					changes = changes + fmt.Sprintf("You can't add any more roles. Limit: %v\n", cons.LimitRoles)
					errfound = true
					continue
				}
			}
		}
	}
	if changes != "" {
		color := disc.DiscGreen
		if errfound {
			color = disc.DiscRed
		}
		embed := []*discordgo.MessageEmbed{{
			Title:       "Changes:",
			Description: changes,
			Color:       color,
		}}
		respose := &discordgo.WebhookEdit{
			Embeds: &embed,
		}
		_, err := s.InteractionResponseEdit(i.Interaction, respose)
		if err != nil {
			cwlog.DoLog("Error: " + err.Error())
		}

		if dirty {
			db.WriteAllCluster()
			db.DumpGuilds()
		}
	}
}
