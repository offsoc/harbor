// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sbom

import (
	"context"

	"github.com/google/uuid"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/sbom/dao"
	"github.com/goharbor/harbor/src/pkg/scan/sbom/model"
)

var (
	// Mgr is the global sbom report manager
	Mgr = NewManager()
)

// Manager is used to manage the sbom reports.
type Manager interface {
	// Create a new report record.
	//
	//  Arguments:
	//    ctx context.Context : the context for this method
	//    r *scan.Report : report model to be created
	//
	//  Returns:
	//    string : uuid of the new report
	//    error  : non nil error if any errors occurred
	//
	Create(ctx context.Context, r *model.Report) (string, error)

	// Delete delete report by uuid
	//
	//  Arguments:
	//    ctx context.Context : the context for this method
	//    uuid string : uuid of the report to delete
	//
	//  Returns:
	//    error  : non nil error if any errors occurred
	//
	Delete(ctx context.Context, uuid string) error

	// UpdateReportData update the report data (with JSON format) of the given report.
	//
	//  Arguments:
	//    ctx context.Context : the context for this method
	//    uuid string    : uuid to identify the report
	//    report string  : report JSON data
	//
	//  Returns:
	//    error  : non nil error if any errors occurred
	//
	UpdateReportData(ctx context.Context, uuid string, report string) error

	// GetBy the reports for the given digest by other properties.
	//
	//  Arguments:
	//    ctx context.Context : the context for this method
	//    artifact_id int64           : the artifact id
	//    registrationUUID string : [optional] the report generated by which registration.
	//                              If it is empty, reports by all the registrations are retrieved.
	//    mimeTypes []string      : [optional] mime types of the reports requiring
	//                              If empty array is specified, reports with all the supported mimes are retrieved.
	//
	//  Returns:
	//    []*Report : sbom report list
	//    error          : non nil error if any errors occurred
	GetBy(ctx context.Context, artifactID int64, registrationUUID string, mimeType string, mediaType string) ([]*model.Report, error)
	// List reports according to the query
	//
	//  Arguments:
	//    ctx context.Context : the context for this method
	//    query *q.Query : the query to list the reports
	//
	//  Returns:
	//    []*scan.Report : report list
	//    error        : non nil error if any errors occurred
	List(ctx context.Context, query *q.Query) ([]*model.Report, error)

	// Update update report information
	Update(ctx context.Context, r *model.Report, cols ...string) error
	// DeleteByExtraAttr delete scan_report by sbom_digest
	DeleteByExtraAttr(ctx context.Context, mimeType, attrName, attrValue string) error
	// DeleteByArtifactID delete sbom report by artifact id
	DeleteByArtifactID(ctx context.Context, artifactID int64) error
}

// basicManager is a default implementation of report manager.
type basicManager struct {
	dao dao.DAO
}

// NewManager news basic manager.
func NewManager() Manager {
	return &basicManager{
		dao: dao.New(),
	}
}

// Create ...
func (bm *basicManager) Create(ctx context.Context, r *model.Report) (string, error) {
	// Validate report object
	if r == nil {
		return "", errors.New("nil sbom report object")
	}

	if r.ArtifactID == 0 || len(r.RegistrationUUID) == 0 || len(r.MimeType) == 0 || len(r.MediaType) == 0 {
		return "", errors.New("malformed sbom report object")
	}

	r.UUID = uuid.New().String()

	// Insert
	if _, err := bm.dao.Create(ctx, r); err != nil {
		return "", err
	}

	return r.UUID, nil
}

func (bm *basicManager) Delete(ctx context.Context, uuid string) error {
	query := q.Query{Keywords: q.KeyWords{"uuid": uuid}}
	count, err := bm.dao.DeleteMany(ctx, query)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.Errorf("no report with uuid %s deleted", uuid)
	}
	return nil
}

// GetBy ...
func (bm *basicManager) GetBy(ctx context.Context, artifactID int64, registrationUUID string,
	mimeType string, mediaType string) ([]*model.Report, error) {
	if artifactID == 0 {
		return nil, errors.New("no artifact id to get sbom report data")
	}

	kws := make(map[string]any)
	kws["artifact_id"] = artifactID
	if len(registrationUUID) > 0 {
		kws["registration_uuid"] = registrationUUID
	}
	if len(mimeType) > 0 {
		kws["mine_type"] = mimeType
	}
	if len(mediaType) > 0 {
		kws["media_type"] = mediaType
	}
	// Query all
	query := &q.Query{
		PageNumber: 0,
		Keywords:   kws,
	}

	return bm.dao.List(ctx, query)
}

// UpdateReportData ...
func (bm *basicManager) UpdateReportData(ctx context.Context, uuid string, report string) error {
	if len(uuid) == 0 {
		return errors.New("missing uuid")
	}

	if len(report) == 0 {
		return errors.New("missing report JSON data")
	}

	return bm.dao.UpdateReportData(ctx, uuid, report)
}

func (bm *basicManager) List(ctx context.Context, query *q.Query) ([]*model.Report, error) {
	return bm.dao.List(ctx, query)
}

func (bm *basicManager) Update(ctx context.Context, r *model.Report, cols ...string) error {
	return bm.dao.Update(ctx, r, cols...)
}

func (bm *basicManager) DeleteByExtraAttr(ctx context.Context, mimeType, attrName, attrValue string) error {
	return bm.dao.DeleteByExtraAttr(ctx, mimeType, attrName, attrValue)
}

func (bm *basicManager) DeleteByArtifactID(ctx context.Context, artifactID int64) error {
	_, err := bm.dao.DeleteMany(ctx, *q.New(q.KeyWords{"ArtifactID": artifactID}))
	return err
}
