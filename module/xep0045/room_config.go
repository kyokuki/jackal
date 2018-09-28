package xep0045

import (
	"github.com/ortuman/jackal/module/xep0004"
	"github.com/ortuman/jackal/xmpp"
)

const (
	MucRoomConfigAnonymityKey             = "muc#roomconfig_anonymity"
	MucRoomConfigChangeSubjectKey         = "muc#roomconfig_changesubject"
	MucRoomConfigEnableLoggingKey         = "muc#roomconfig_enablelogging"
	MucRoomConfigMaxHistoryKey            = "muc#maxhistoryfetch"
	MucRoomConfigMaxUsersKey              = "muc#roomconfig_maxusers"
	MucRoomConfigMaxUserResourcesKey      = "muc#roomconfig_maxresources"
	MucRoomConfigMembersOnlyKey           = "muc#roomconfig_membersonly"
	MucRoomConfigAllowInvitesKey          = "muc#roomconfig_allowinvites"
	MucRoomConfigModeratedRoomKey         = "muc#roomconfig_moderatedroom"
	MucRoomConfigPasswordProtectedRoomKey = "muc#roomconfig_passwordprotectedroom"
	MucRoomConfigPersistentRoomKey        = "muc#roomconfig_persistentroom"
	MucRoomConfigPublicRoomKey            = "muc#roomconfig_publicroom"
	MucRoomConfigRoomDescKey              = "muc#roomconfig_roomdesc"
	MucRoomConfigRoomNameKey              = "muc#roomconfig_roomname"
	MucRoomConfigRoomSecretKey            = "muc#roomconfig_roomsecret"
)

type RoomConfig struct {
	init bool
	roomJID string
	form *xep0004.Form
}

func NewRoomConfig() RoomConfig {
	return RoomConfig{
		init : true,
		roomJID: "",
		form : newRoomConfigForm(),
	}
}

func newRoomConfigForm() *xep0004.Form {
	form := xep0004.New("form", "", "")
	form.AddField(xep0004.NewFieldTextSingle(MucRoomConfigRoomNameKey, "", "Natural-Language Room Name"))
	form.AddField(xep0004.NewFieldTextSingle(MucRoomConfigRoomDescKey, "", "Short Description of Room"))
	form.AddField(xep0004.NewFieldBool(MucRoomConfigPersistentRoomKey, false, "Make Room Persistent?"))
	form.AddField(xep0004.NewFieldBool(MucRoomConfigPublicRoomKey, true, "Make Room Publicly Searchable?"))
	form.AddField(xep0004.NewFieldBool(MucRoomConfigModeratedRoomKey, false, "Make Room Moderated?"))
	form.AddField(xep0004.NewFieldBool(MucRoomConfigMembersOnlyKey, false, "Make Room Members Only?"))
	form.AddField(xep0004.NewFieldBool(MucRoomConfigAllowInvitesKey, true, "Allow Occupants to Invite Others?"))
	form.AddField(xep0004.NewFieldBool(MucRoomConfigPasswordProtectedRoomKey, false, "Password Required to Enter?"))
	form.AddField(xep0004.NewFieldTextSingle(MucRoomConfigRoomSecretKey, "", "Password"))
	form.AddField(xep0004.NewFieldBool(MucRoomConfigChangeSubjectKey, false, "Allow Occupants to Change Subject?"))
	form.AddField(xep0004.NewFieldBool(MucRoomConfigEnableLoggingKey, false, "Enable Public Logging?"))
	form.AddField(xep0004.NewFieldTextSingle(MucRoomConfigMaxHistoryKey, "50", "Maximum Number of History Messages Returned by Room"))

	fieldMucRoomConfigMaxUsersKey, err := xep0004.NewFieldListSingle(MucRoomConfigMaxUsersKey, "", "Maximum Number of Occupants",
		[]string{"10", "20", "30", "50", "100", "None"},
		[]string{"10", "20", "30", "50", "100", ""},
	)
	if err == nil {
		form.AddField(fieldMucRoomConfigMaxUsersKey)
	}

	fieldMucRoomConfigMaxUserResourcesKey, err := xep0004.NewFieldListSingle(MucRoomConfigMaxUserResourcesKey, "", "Maximum Number of Single Occupant Resources",
		[]string{"5", "10", "20", "30", "50", "100", "None"},
		[]string{"5", "10", "20", "30", "50", "100", ""},
	)
	if err == nil {
		form.AddField(fieldMucRoomConfigMaxUserResourcesKey)
	}

	return form
}

func (rc RoomConfig) AsElement() *xmpp.Element {
	if !rc.init {
		return nil
	}
	return rc.form.Element()
}