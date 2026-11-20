package novelcreation

import _ "embed"

// Native source: novels.prompts.inkos.short_fiction_prompts constants72-80,
// novels.pyd b5b22086, byte ranges4238903..4239568. The native builder writes a
// complete second draft in Markdown. This version explicitly adapts its review
// priorities to CM's current chapter/piece JSON contract; it is not that builder.
const ShortWriterRevisionFollowupID = "novel-creation.short-writer-revision-followup"

//go:embed short_writer/revision_followup.txt
var ShortWriterRevisionFollowupSystem string
