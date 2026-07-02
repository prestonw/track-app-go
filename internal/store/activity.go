package store

import (
	"os"
	"path/filepath"

	"github.com/prestonw/track-app-go/internal/models"
)

func (s *Store) OpenActivitySegment() *models.ActivitySegment {
	for i := range s.ActivityLog {
		if s.ActivityLog[i].IsOpen() {
			return &s.ActivityLog[i]
		}
	}
	return nil
}

func (s *Store) StartActivitySegment(ctx models.ForegroundContext, projectID string) {
	if open := s.OpenActivitySegment(); open != nil {
		if open.AppName == ctx.AppName && open.BundleID == ctx.BundleID &&
			open.WindowTitle == ctx.WindowTitle && open.DocumentPath == ctx.DocumentPath &&
			open.ProjectID == projectID {
			return
		}
		s.CloseOpenActivitySegment()
	}
	seg := models.ActivitySegment{
		ID: models.MakeID(), StartedAt: models.NowMs(),
		AppName: ctx.AppName, BundleID: ctx.BundleID,
		WindowTitle: ctx.WindowTitle, DocumentPath: ctx.DocumentPath,
		ProjectID: projectID,
	}
	s.ActivityLog = append([]models.ActivitySegment{seg}, s.ActivityLog...)
	s.persistActivitySegment(seg)
}

func (s *Store) CloseOpenActivitySegment() {
	open := s.OpenActivitySegment()
	if open == nil {
		return
	}
	now := models.NowMs()
	for i := range s.ActivityLog {
		if s.ActivityLog[i].ID == open.ID {
			s.ActivityLog[i].EndedAt = &now
			s.updateActivitySegment(s.ActivityLog[i])
			return
		}
	}
}

func (s *Store) persistActivitySegment(seg models.ActivitySegment) {
	_, _ = s.db.Exec(`INSERT INTO activity_log VALUES (?,?,?,?,?,?,?,?)`,
		seg.ID, seg.StartedAt, "", seg.AppName, seg.BundleID, seg.WindowTitle, seg.DocumentPath, seg.ProjectID)
}

func (s *Store) updateActivitySegment(seg models.ActivitySegment) {
	ended := ""
	if seg.EndedAt != nil {
		ended = itoa64(*seg.EndedAt)
	}
	_, _ = s.db.Exec(`UPDATE activity_log SET ended_at=?, project_id=? WHERE id=?`,
		ended, seg.ProjectID, seg.ID)
}

func (s *Store) ExportDatabase(dest string) error {
	src := filepath.Join(s.dir, "track-app.sqlite")
	if s.db != nil {
		_ = s.SaveAll()
	}
	return copyFile(src, dest)
}

func copyFile(src, dest string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, b, 0o644)
}