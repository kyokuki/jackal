/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package host

import (
	"os"
	"testing"

	"github.com/ortuman/jackal/util"
	"github.com/stretchr/testify/require"
)

func TestHost(t *testing.T) {
	hm, _ := New(nil)
	Init(hm)
	require.True(t, IsLocalHost("localhost"))
	require.False(t, IsLocalHost("jackal.im"))
	require.Equal(t, 1, len(HostNames()))
	os.RemoveAll("./.cert")
	Close()

	hm, _ = New([]Config{{Name: "jackal.im"}})
	Init(hm)
	require.False(t, IsLocalHost("localhost"))
	require.True(t, IsLocalHost("jackal.im"))
	require.Equal(t, 1, len(HostNames()))
	Close()

	privKeyFile := "../testdata/cert/test.server.key"
	certFile := "../testdata/cert/test.server.crt"
	cer, err := util.LoadCertificate(privKeyFile, certFile, "localhost")
	require.Nil(t, err)

	hm, _ = New([]Config{{Name: "localhost", Certificate: cer}})
	Init(hm)
	require.Equal(t, 1, len(Certificates()))
	require.Equal(t, 1, len(HostNames()))
	Close()
}
