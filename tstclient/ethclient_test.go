// Copyright 2016 The go-utchain Authors
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

package tstclient

import "github.com/utchain/go-utchain"

// Verify that Client implements the utereum interfaces.
var (
	_ = utereum.ChainReader(&Client{})
	_ = utereum.TransactionReader(&Client{})
	_ = utereum.ChainStateReader(&Client{})
	_ = utereum.ChainSyncReader(&Client{})
	_ = utereum.ContractCaller(&Client{})
	_ = utereum.GasEstimator(&Client{})
	_ = utereum.GasPricer(&Client{})
	_ = utereum.LogFilterer(&Client{})
	_ = utereum.PendingStateReader(&Client{})
	// _ = utereum.PendingStateEventer(&Client{})
	_ = utereum.PendingContractCaller(&Client{})
)
