package skills

import _ "embed"

//go:embed kiosk/SKILL.md
var KioskSkill string

//go:embed kiosk/init-prompt.md
var KioskInitPrompt string

//go:embed kiosk/publish-prompt.md
var KioskPublishPrompt string
