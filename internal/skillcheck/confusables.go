package skillcheck

// ConfusableMap maps visually confusable characters to their Latin equivalents.
// Sourced from Unicode confusables.txt — focused on the most dangerous pairs
// used in homoglyph attacks (Cyrillic/Greek characters that look like Latin).
var ConfusableMap = map[rune]rune{
	// Cyrillic → Latin confusables
	'\u0410': 'A', // А → A
	'\u0412': 'B', // В → B
	'\u0421': 'C', // С → C
	'\u0415': 'E', // Е → E
	'\u041D': 'H', // Н → H
	'\u0406': 'I', // І → I
	'\u0408': 'J', // Ј → J
	'\u041A': 'K', // К → K
	'\u041C': 'M', // М → M
	'\u041E': 'O', // О → O
	'\u0420': 'P', // Р → P
	'\u0405': 'S', // Ѕ → S
	'\u0422': 'T', // Т → T
	'\u0425': 'X', // Х → X
	'\u0430': 'a', // а → a
	'\u0435': 'e', // е → e
	'\u043E': 'o', // о → o
	'\u0440': 'p', // р → p
	'\u0441': 'c', // с → c
	'\u0445': 'x', // х → x
	'\u0455': 's', // ѕ → s
	'\u0456': 'i', // і → i
	'\u0458': 'j', // ј → j
	'\u04BB': 'h', // һ → h
	'\u04C0': 'I', // Ӏ → I
	'\u04CF': 'l', // ӏ → l

	// Greek → Latin confusables
	'\u0391': 'A', // Α → A
	'\u0392': 'B', // Β → B
	'\u0395': 'E', // Ε → E
	'\u0396': 'Z', // Ζ → Z
	'\u0397': 'H', // Η → H
	'\u0399': 'I', // Ι → I
	'\u039A': 'K', // Κ → K
	'\u039C': 'M', // Μ → M
	'\u039D': 'N', // Ν → N
	'\u039F': 'O', // Ο → O
	'\u03A1': 'P', // Ρ → P
	'\u03A4': 'T', // Τ → T
	'\u03A5': 'Y', // Υ → Y
	'\u03A7': 'X', // Χ → X
	'\u03B1': 'a', // α → a (less confusable but included)
	'\u03BF': 'o', // ο → o
	'\u03C1': 'p', // ρ → p

	// Other confusables
	'\u0131': 'i', // ı (dotless i) → i
	'\u0237': 'j', // ȷ (dotless j) → j
	'\u1D00': 'A', // ᴀ (small capital A)
	'\u1D04': 'C', // ᴄ (small capital C)
	'\u1D05': 'D', // ᴅ (small capital D)
	'\u1D07': 'E', // ᴇ (small capital E)
	'\u1D0A': 'J', // ᴊ (small capital J)
	'\u1D0B': 'K', // ᴋ (small capital K)
	'\u1D0D': 'M', // ᴍ (small capital M)
	'\u1D0F': 'O', // ᴏ (small capital O)
	'\u1D18': 'P', // ᴘ (small capital P)
	'\u1D1B': 'T', // ᴛ (small capital T)
	'\u1D1C': 'U', // ᴜ (small capital U)
	'\u1D20': 'V', // ᴠ (small capital V)
	'\u1D21': 'W', // ᴡ (small capital W)
	'\u1D22': 'Z', // ᴢ (small capital Z)

	// Fullwidth Latin
	'\uFF21': 'A', // Ａ
	'\uFF22': 'B', // Ｂ
	'\uFF23': 'C', // Ｃ
	'\uFF24': 'D', // Ｄ
	'\uFF25': 'E', // Ｅ
	'\uFF26': 'F', // Ｆ
	'\uFF27': 'G', // Ｇ
	'\uFF28': 'H', // Ｈ
	'\uFF29': 'I', // Ｉ
	'\uFF2A': 'J', // Ｊ
	'\uFF2B': 'K', // Ｋ
	'\uFF2C': 'L', // Ｌ
	'\uFF2D': 'M', // Ｍ
	'\uFF2E': 'N', // Ｎ
	'\uFF2F': 'O', // Ｏ
	'\uFF30': 'P', // Ｐ
	'\uFF31': 'Q', // Ｑ
	'\uFF32': 'R', // Ｒ
	'\uFF33': 'S', // Ｓ
	'\uFF34': 'T', // Ｔ
	'\uFF35': 'U', // Ｕ
	'\uFF36': 'V', // Ｖ
	'\uFF37': 'W', // Ｗ
	'\uFF38': 'X', // Ｘ
	'\uFF39': 'Y', // Ｙ
	'\uFF3A': 'Z', // Ｚ
	'\uFF41': 'a', // ａ
	'\uFF42': 'b', // ｂ
	'\uFF43': 'c', // ｃ
	'\uFF44': 'd', // ｄ
	'\uFF45': 'e', // ｅ
	'\uFF46': 'f', // ｆ
	'\uFF47': 'g', // ｇ
	'\uFF48': 'h', // ｈ
	'\uFF49': 'i', // ｉ
	'\uFF4A': 'j', // ｊ
	'\uFF4B': 'k', // ｋ
	'\uFF4C': 'l', // ｌ
	'\uFF4D': 'm', // ｍ
	'\uFF4E': 'n', // ｎ
	'\uFF4F': 'o', // ｏ
	'\uFF50': 'p', // ｐ
	'\uFF51': 'q', // ｑ
	'\uFF52': 'r', // ｒ
	'\uFF53': 's', // ｓ
	'\uFF54': 't', // ｔ
	'\uFF55': 'u', // ｕ
	'\uFF56': 'v', // ｖ
	'\uFF57': 'w', // ｗ
	'\uFF58': 'x', // ｘ
	'\uFF59': 'y', // ｙ
	'\uFF5A': 'z', // ｚ
}
