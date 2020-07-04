// Copyright (C) 2020 Felix Geyer <debfx@fobos.de>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 2 or (at your option)
// version 3 of the License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type jailProfile struct {
	EnvVariables map[string]string
	Mounts       []mount
}

func getX11Socket() (string, error) {
	display := os.Getenv("DISPLAY")

	if len(display) == 0 {
		return "", fmt.Errorf("DISPLAY envrionment variable is not set")
	}

	if len(display) < 2 || display[0] != ':' {
		return "", fmt.Errorf("DISPLAY envrionment variable is invalid (\"%s\")", display)
	}

	if _, err := strconv.Atoi(display[1:]); err != nil {
		return "", fmt.Errorf("DISPLAY envrionment variable is invalid (\"%s\")", display)
	}

	socketPath := fmt.Sprintf("/tmp/.X11-unix/X%s", display[1:])

	if _, err := os.Stat(socketPath); err != nil {
		return "", err
	}

	return socketPath, nil
}

func getWaylandSocket() (string, error) {
	socketPath := os.Getenv("WAYLAND_DISPLAY")
	if len(socketPath) == 0 {
		socketPath = "wayland-0"
	}

	if !strings.HasPrefix(socketPath, "/") {
		runtimeDir, err := getUserRuntimeDir()
		if err != nil {
			return "", err
		}

		socketPath = fmt.Sprintf("%s/%s", runtimeDir, socketPath)
	}

	if _, err := os.Stat(socketPath); err != nil {
		return "", err
	}

	return socketPath, nil
}

func getProfile(name string) (jailProfile, error) {
	profile := jailProfile{}

	switch name {
	case "x11":
		x11Socket, err := getX11Socket()
		if err != nil {
			return profile, err
		}

		profile.Mounts = append(profile.Mounts, mount{Path: "/tmp/.X11-unix/X0", Other: x11Socket, Type: mountTypeBindRw})
		profile.EnvVariables["DISPLAY"] = ":0"
	case "wayland":
		waylandSocket, err := getWaylandSocket()
		if err != nil {
			return profile, err
		}

		runtimeDir, err := getUserRuntimeDir()
		if err != nil {
			return profile, err
		}

		profile.Mounts = append(profile.Mounts, mount{Path: fmt.Sprintf("%s/wayland-0", runtimeDir), Other: waylandSocket, Type: mountTypeBindRw})
		profile.EnvVariables["WAYLAND_DISPLAY"] = "wayland-0"
	}

	return profile, nil
}
