// Copyright © 2020 Cisco
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// All rights reserved.

package servicedirectory

import (
	"context"
	"strings"
	"time"

	sr "github.com/CloudNativeSDWAN/cnwan-operator/pkg/servregistry"
	"google.golang.org/api/iterator"
	sdpb "google.golang.org/genproto/googleapis/cloud/servicedirectory/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetServ returns the service if exists.
func (s *servDir) GetServ(nsName, servName string) (*sr.Service, error) {
	// -- Init
	if err := s.checkNames(&nsName, &servName, nil); err != nil {
		return nil, err
	}
	l := s.log.WithName("GetServ").WithValues("ns-name", nsName, "serv-name", servName)
	servPath := s.getResourcePath(servDirPath{namespace: nsName, service: servName})
	ctx, canc := context.WithTimeout(s.context, s.timeout)
	defer canc()

	sdServ, err := s.client.GetService(ctx, &sdpb.GetServiceRequest{Name: servPath})
	if err == nil {
		serv := &sr.Service{
			Name:     servName,
			NsName:   nsName,
			Metadata: sdServ.Annotations,
		}
		if serv.Metadata == nil {
			serv.Metadata = map[string]string{}
		}

		return serv, nil
	}

	// What is the error?
	if err == context.DeadlineExceeded {
		l.Error(err, "timeout expired while waiting for service directory to reply", "timeout-seconds", s.timeout.Seconds())
		return nil, sr.ErrTimeOutExpired
	}

	if status.Code(err) == codes.NotFound {
		return nil, sr.ErrNotFound
	}

	// Any other error
	return nil, err
}

// ListServ returns a list of services inside the provided namespace.
func (s *servDir) ListServ(nsName string) (servList []*sr.Service, err error) {
	// -- Init
	if err := s.checkNames(&nsName, nil, nil); err != nil {
		return nil, err
	}
	l := s.log.WithName("ListServ").WithValues("ns-name", nsName)
	ctx, canc := context.WithTimeout(s.context, time.Minute)
	defer canc()

	req := &sdpb.ListServicesRequest{
		Parent: s.getResourcePath(servDirPath{namespace: nsName}),
	}

	iter := s.client.ListServices(ctx, req)
	if iter == nil {
		l.V(0).Info("returned list is nil")
		return
	}
	for {
		nextServ, iterErr := iter.Next()
		if iterErr != nil {

			if iterErr == context.DeadlineExceeded {
				l.Error(err, "timeout expired while waiting for service directory to reply", "timeout-seconds", s.timeout.Seconds())
				return nil, sr.ErrTimeOutExpired
			}

			if iterErr != iterator.Done {
				l.Error(iterErr, "error while loading services")
				return nil, iterErr
			}

			break
		}

		// Create the list
		splitName := strings.Split(nextServ.Name, "/")
		serv := &sr.Service{
			Name:     splitName[len(splitName)-1],
			NsName:   nsName,
			Metadata: nextServ.Annotations,
		}
		if serv.Metadata == nil {
			serv.Metadata = map[string]string{}
		}

		servList = append(servList, serv)
	}

	return
}

// CreateServ creates the service.
func (s *servDir) CreateServ(serv *sr.Service) (*sr.Service, error) {
	// -- Init
	if serv == nil {
		return nil, sr.ErrServNotProvided
	}
	if err := s.checkNames(&serv.NsName, &serv.Name, nil); err != nil {
		return nil, err
	}
	l := s.log.WithName("CreateServ").WithValues("ns-name", serv.NsName, "serv-name", serv.Name, "metadata", serv.Metadata)
	ctx, canc := context.WithTimeout(s.context, s.timeout)
	defer canc()

	servToCreate := &sdpb.Service{
		Name:        serv.Name,
		Annotations: serv.Metadata,
	}

	req := &sdpb.CreateServiceRequest{
		Parent:    s.getResourcePath(servDirPath{namespace: serv.NsName}),
		ServiceId: serv.Name,
		Service:   servToCreate,
	}

	_, err := s.client.CreateService(ctx, req)
	if err == nil {
		// If it is successful, then it makes no point in parsing the returned
		// service from service directory, because it will look like just the
		// same as the service we want to create, apart from having prefixes
		// in the name, which is something we want to abstract to someone
		// using this.
		return serv, nil
	}

	// What is the error?
	if err == context.DeadlineExceeded {
		l.Error(err, "timeout expired while waiting for service directory to reply", "timeout-seconds", s.timeout.Seconds())
		return nil, sr.ErrTimeOutExpired
	}

	if status.Code(err) == codes.AlreadyExists {
		return nil, sr.ErrAlreadyExists
	}

	// Any other error
	return nil, err
}

// UpdateServ updates the service.
func (s *servDir) UpdateServ(serv *sr.Service) (*sr.Service, error) {
	// -- Init
	if serv == nil {
		return nil, sr.ErrServNotProvided
	}
	if err := s.checkNames(&serv.NsName, &serv.Name, nil); err != nil {
		return nil, err
	}
	l := s.log.WithName("UpdateServ").WithValues("ns-name", serv.NsName, "serv-name", serv.Name, "metadata", serv.Metadata)
	ctx, canc := context.WithTimeout(s.context, s.timeout)
	defer canc()

	servToUpd := &sdpb.Service{
		Name:        s.getResourcePath(servDirPath{namespace: serv.NsName, service: serv.Name}),
		Annotations: serv.Metadata,
	}

	req := &sdpb.UpdateServiceRequest{
		Service: servToUpd,
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"metadata"},
		},
	}

	_, err := s.client.UpdateService(ctx, req)
	if err == nil {
		return serv, nil
	}

	// What is the error?
	if err == context.DeadlineExceeded {
		l.Error(err, "timeout expired while waiting for service directory to reply", "timeout-seconds", s.timeout.Seconds())
		return nil, sr.ErrTimeOutExpired
	}

	if status.Code(err) == codes.NotFound {
		return nil, sr.ErrNotFound
	}

	// Any other error
	return nil, err
}

// DeleteServ deletes the service.
func (s *servDir) DeleteServ(nsName, servName string) error {
	// -- Init
	if err := s.checkNames(&nsName, &servName, nil); err != nil {
		return err
	}
	l := s.log.WithName("DeleteServ").WithValues("ns-name", nsName, "serv-name", servName)
	ctx, canc := context.WithTimeout(s.context, s.timeout)
	defer canc()

	req := &sdpb.DeleteServiceRequest{
		Name: s.getResourcePath(servDirPath{namespace: nsName, service: servName}),
	}

	err := s.client.DeleteService(ctx, req)
	if err == nil {
		return nil
	}

	// What is the error?
	if err == context.DeadlineExceeded {
		l.Error(err, "timeout expired while waiting for service directory to reply", "timeout-seconds", s.timeout.Seconds())
		return sr.ErrTimeOutExpired
	}

	if status.Code(err) == codes.NotFound {
		return sr.ErrNotFound
	}

	// Any other error
	return err
}
