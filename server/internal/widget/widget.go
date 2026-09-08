// Package widget serves the embeddable auth widget script.
package widget

import _ "embed"

// EmbedJS is the standalone widget loader for third-party sites.
//
//go:embed embed.js
var EmbedJS []byte
