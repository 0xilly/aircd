package libpbn

var (
	UNICODE_COLOR_START rune = 0x100000
	UNICODE_COLOR_END   rune = 0x10FFFF
	UNICODE_INDEX_START rune = 0x02FA1E
	UNICODE_RUN_MARK    rune = 0x00FFFF
	RUNES_PER_INDEX     int  = 1
	RUNES_PER_COLOR     int  = 2
	RUNES_PER_RUN       int  = 3
)
