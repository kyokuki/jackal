/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package host

import "crypto/tls"

type disabledHostManager struct{}

func (_ *disabledHostManager) HostNames() []string             { return nil }
func (_ *disabledHostManager) IsLocalHost(domain string) bool  { return false }
func (_ *disabledHostManager) Certificates() []tls.Certificate { return nil }
func (_ *disabledHostManager) Close()                          {}
