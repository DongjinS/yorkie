/*
 * Copyright 2022 The Yorkie Authors. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package grpchelper

import (
	"errors"
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yorkie-team/yorkie/api/converter"
	"github.com/yorkie-team/yorkie/api/types"
	"github.com/yorkie-team/yorkie/pkg/document/time"
	"github.com/yorkie-team/yorkie/server/backend/database"
	"github.com/yorkie-team/yorkie/server/clients"
	"github.com/yorkie-team/yorkie/server/packs"
	"github.com/yorkie-team/yorkie/server/rpc/auth"
)

// ToStatusError returns a status.Error from the given logic error. If an error
// occurs while executing logic in API handler, gRPC status.error should be
// returned so that the client can know more about the status of the request.
func ToStatusError(err error) error {
	var st *status.Status
	// TODO(DONGJIN SHIN): Get details from validation
	var details interface{}
	if errors.Is(err, auth.ErrNotAllowed) ||
		errors.Is(err, auth.ErrUnexpectedStatusCode) ||
		errors.Is(err, auth.ErrWebhookTimeout) {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	if errors.Is(err, converter.ErrPackRequired) ||
		errors.Is(err, converter.ErrCheckpointRequired) ||
		errors.Is(err, time.ErrInvalidHexString) ||
		errors.Is(err, time.ErrInvalidActorID) ||
		errors.Is(err, types.ErrInvalidID) ||
		errors.Is(err, clients.ErrInvalidClientID) ||
		errors.Is(err, clients.ErrInvalidClientKey) ||
		errors.Is(err, types.ErrEmptyProjectFields) ||
		errors.Is(err, types.ErrNotSupportedMethod) ||
		errors.Is(err, types.ErrInvalidProjectField) {
		st = status.New(codes.InvalidArgument, err.Error())
	}

	if errors.Is(err, converter.ErrUnsupportedOperation) ||
		errors.Is(err, converter.ErrUnsupportedElement) ||
		errors.Is(err, converter.ErrUnsupportedEventType) ||
		errors.Is(err, converter.ErrUnsupportedValueType) ||
		errors.Is(err, converter.ErrUnsupportedCounterType) {
		return status.Error(codes.Unimplemented, err.Error())
	}

	if errors.Is(err, database.ErrProjectNotFound) ||
		errors.Is(err, database.ErrClientNotFound) ||
		errors.Is(err, database.ErrDocumentNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}

	if errors.Is(err, database.ErrProjectAlreadyExists) ||
		errors.Is(err, database.ErrProjectNameAlreadyExists) {
		return status.Error(codes.AlreadyExists, err.Error())
	}

	if err == database.ErrClientNotActivated ||
		err == database.ErrDocumentNotAttached ||
		err == database.ErrDocumentAlreadyAttached ||
		errors.Is(err, packs.ErrInvalidServerSeq) ||
		errors.Is(err, database.ErrConflictOnUpdate) {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	if st != nil {
		if details != nil {
			st, _ = st.WithDetails(details.(*errdetails.BadRequest))
			if err != nil {
				panic(fmt.Sprintf("Unexpected error: %v", err))
			}
		}
		return st.Err()
	}

	return status.Error(codes.Internal, err.Error())
}
