/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package component

import (
	"github.com/ortuman/jackal/component/httpupload"
	"github.com/ortuman/jackal/component/pubsub"
)

// Config contains all components configuration.
type Config struct {
	HttpUpload *httpupload.Config `yaml:"http_upload"`
	Pubsub     *pubsub.Config     `yaml:"pubsub"`
}
