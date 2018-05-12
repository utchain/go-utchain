// Copyright 2018 The go-utchain Authors
// This file is part of the go-utchain library.
//
// The go-utchain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-utchain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-utchain library. If not, see <http://www.gnu.org/licenses/>.

package dashboard

// getProcessCPUTime returns 0 on Windows as there is no system call to resolve
// the actual process' CPU time.
func getProcessCPUTime() float64 {
	return 0
}
