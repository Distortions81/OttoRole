package disc

import (
	"RoleKeeper/cwlog"

	"github.com/bwmarrin/discordgo"
)

var (
	Session *discordgo.Session
	Ready   *discordgo.Ready
)

const (
	DiscRed    = 0xFF0000
	DiscOrange = 0xFFFF00
	DiscGreen  = 0x00FF00
	DiscCyan   = 0x00FFFF
	DiscBlue   = 0x0000FF
	DiscPurple = 0xFF00FF
)

func GetGuildRoles(s *discordgo.Session, guildid string) []*discordgo.Role {
	guild, err := s.Guild(guildid)
	if guild != nil && err == nil {
		return guild.Roles
	}
	return nil
}

func InteractionResponse(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	cwlog.DoLog("InteractionResponse:\n" + i.Member.User.Username + "\n" + embed.Title + "\n" + embed.Description)

	var embedList []*discordgo.MessageEmbed
	embedList = append(embedList, embed)
	respData := &discordgo.InteractionResponseData{Embeds: embedList}
	resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
	err := s.InteractionRespond(i.Interaction, resp)
	if err != nil {
		cwlog.DoLog(err.Error())
	}
}

func FollowupResponse(s *discordgo.Session, i *discordgo.InteractionCreate, f *discordgo.WebhookParams) {
	if f.Embeds != nil {
		cwlog.DoLog("FollowupResponse:\n" + i.Member.User.Username + "\n" + f.Embeds[0].Title + "\n" + f.Embeds[0].Description)

		_, err := s.FollowupMessageCreate(i.Interaction, false, f)
		if err != nil {
			cwlog.DoLog(err.Error())
		}
	} else if f.Content != "" {
		cwlog.DoLog("FollowupResponse:\n" + i.Member.User.Username + "\n" + f.Content)

		_, err := s.FollowupMessageCreate(i.Interaction, false, f)
		if err != nil {
			cwlog.DoLog(err.Error())
		}
	}

}

func EphemeralResponse(s *discordgo.Session, i *discordgo.InteractionCreate, color int, title, message string) {
	cwlog.DoLog("EphemeralResponse:\n" + i.Member.User.Username + "\n" + title + "\n" + message)

	var elist []*discordgo.MessageEmbed
	elist = append(elist, &discordgo.MessageEmbed{Title: title, Description: message, Color: color})

	respData := &discordgo.InteractionResponseData{Embeds: elist, Flags: discordgo.MessageFlagsEphemeral}
	resp := &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: respData}
	err := s.InteractionRespond(i.Interaction, resp)
	if err != nil {
		cwlog.DoLog(err.Error())
	}
}
