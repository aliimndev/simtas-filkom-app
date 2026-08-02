package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	domainRepo "github.com/aliimndev/simtas-filkom-app/backend/internal/domain/repository"
)

type documentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) domainRepo.DocumentRepository {
	return &documentRepository{db: db}
}

// preloadDocument applies the association preloads used for document reads.
func preloadDocument(q *gorm.DB) *gorm.DB {
	return q.Preload("Reviewer.Role")
}

func (r *documentRepository) Create(ctx context.Context, doc *entity.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *documentRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Document, error) {
	var doc entity.Document
	err := preloadDocument(r.db.WithContext(ctx)).
		Where("id = ?", id).
		First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// baseQuery builds the shared thesis+filter scope used by every document query.
func (r *documentRepository) baseQuery(ctx context.Context, thesisID uuid.UUID, filter domainRepo.DocumentFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&entity.Document{}).
		Where("documents.thesis_id = ?", thesisID)
	if filter.DocumentType != "" {
		q = q.Where("documents.document_type = ?", filter.DocumentType)
	}
	if filter.Status != "" {
		q = q.Where("documents.status = ?", filter.Status)
	}
	return q
}

// FindByThesisID returns the active documents (latest version per
// document_type/chapter) of a thesis, newest first.
//
// It first resolves the latest version IDs with DISTINCT ON, then fetches the
// rows with association preloads — this keeps GORM preloads working (they
// reference the documents table, not an aliased subquery).
func (r *documentRepository) FindByThesisID(ctx context.Context, thesisID uuid.UUID, filter domainRepo.DocumentFilter) ([]*entity.Document, int64, error) {
	base := r.baseQuery(ctx, thesisID, filter)

	// Total = number of distinct active (document_type, chapter) pairs.
	var total int64
	if err := base.Session(&gorm.Session{}).
		Distinct("document_type, COALESCE(chapter_number, 0)").
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, perPage := filter.Page, filter.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	// Resolve the latest version id per (document_type, chapter), paginated.
	var rows []struct {
		ID uuid.UUID
	}
	err := base.Session(&gorm.Session{}).
		Select("DISTINCT ON (document_type, COALESCE(chapter_number, 0)) documents.id").
		Order("document_type, COALESCE(chapter_number, 0), version DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return []*entity.Document{}, total, nil
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}

	// Fetch with preloads, newest uploads first.
	var docs []*entity.Document
	if err := preloadDocument(r.db.WithContext(ctx)).
		Where("documents.id IN ?", ids).
		Order("documents.created_at DESC").
		Find(&docs).Error; err != nil {
		return nil, 0, err
	}
	return docs, total, nil
}

func (r *documentRepository) FindLatestByType(ctx context.Context, thesisID uuid.UUID, docType string, chapterNum *int) (*entity.Document, error) {
	q := r.db.WithContext(ctx).
		Where("thesis_id = ? AND document_type = ?", thesisID, docType)
	if docType == entity.DocTypeDraftChapter && chapterNum != nil {
		q = q.Where("chapter_number = ?", *chapterNum)
	}

	var doc entity.Document
	err := q.Order("version DESC").First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *documentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, reviewerID uuid.UUID, notes string) error {
	updates := map[string]interface{}{
		"status":      status,
		"reviewer_id": reviewerID,
		"reviewed_at": time.Now(),
	}
	if notes != "" {
		updates["reviewer_notes"] = notes
	}
	// Guard the transition so two concurrent reviews cannot both succeed.
	result := r.db.WithContext(ctx).
		Model(&entity.Document{}).
		Where("id = ? AND status = ?", id, entity.DocStatusPendingReview).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *documentRepository) GetVersionHistory(ctx context.Context, thesisID uuid.UUID, docType string, chapterNum *int) ([]*entity.Document, error) {
	q := r.db.WithContext(ctx).
		Where("thesis_id = ? AND document_type = ?", thesisID, docType)
	if docType == entity.DocTypeDraftChapter && chapterNum != nil {
		q = q.Where("chapter_number = ?", *chapterNum)
	}

	var docs []*entity.Document
	err := preloadDocument(q).
		Order("version DESC").
		Find(&docs).Error
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *documentRepository) IsDocumentApproved(ctx context.Context, thesisID uuid.UUID, docType string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Document{}).
		Where("thesis_id = ? AND document_type = ? AND status = ?",
			thesisID, docType, entity.DocStatusApproved).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
