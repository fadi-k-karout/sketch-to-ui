package uicomponents

import (
	"database/sql"
	"fmt"
)

type UIComponentsStore struct {
	db *sql.DB
}

func NewUIComponentsStore(db *sql.DB) *UIComponentsStore {
	return &UIComponentsStore{
		db: db,
	}
}

// CreateComponent creates a new UI component
func (cs *UIComponentsStore) CreateComponent(component *UIComponent) error {
	sqlQuery := `
		INSERT INTO uicomponents (title, type, code, is_public, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at`

	err := cs.db.QueryRow(sqlQuery, component.Title, component.Type, component.Code, component.IsPublic, component.UserID).
		Scan(&component.ID, &component.CreatedAt, &component.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create component: %w", err)
	}

	return nil
}

// UpdateComponent updates an existing UI component
func (cs *UIComponentsStore) UpdateComponent(id int, component *UIComponent) error {
	sqlQuery := `
		UPDATE uicomponents 
		SET title = $1, type = $2, code = $3, is_public = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5 AND archived_at IS NULL
		RETURNING updated_at`

	err := cs.db.QueryRow(sqlQuery, component.Title, component.Type, component.Code, component.IsPublic, id).
		Scan(&component.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("component with id %d not found or already archived", id)
		}
		return fmt.Errorf("failed to update component: %w", err)
	}

	component.ID = id
	return nil
}

// ArchiveComponent soft deletes a UI component by setting archived_at timestamp
func (cs *UIComponentsStore) ArchiveComponent(id int) error {
	sqlQuery := `
		UPDATE uicomponents 
		SET archived_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND archived_at IS NULL`

	result, err := cs.db.Exec(sqlQuery, id)
	if err != nil {
		return fmt.Errorf("failed to archive component: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("component with id %d not found or already archived", id)
	}

	return nil
}

// GetComponentByID retrieves a component by its ID
func (cs *UIComponentsStore) GetComponentByID(id int) (*UIComponent, error) {
	sqlQuery := `
		SELECT id, title, type, code, is_public, user_id, created_at, updated_at
		FROM uicomponents
		WHERE id = $1 AND archived_at IS NULL`

	var component UIComponent
	err := cs.db.QueryRow(sqlQuery, id).Scan(
		&component.ID,
		&component.Title,
		&component.Type,
		&component.Code,
		&component.IsPublic,
		&component.UserID,
		&component.CreatedAt,
		&component.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("component with id %d not found or archived", id)
		}
		return nil, fmt.Errorf("failed to get component by id: %w", err)
	}

	return &component, nil
}

// GetAllComponentsByUser retrieves all non-archived UI components for a user
func (cs *UIComponentsStore) GetAllComponentsByUser(userID int) ([]UIComponent, error) {
	sqlQuery := `
		SELECT id, title, type, code, is_public, user_id, created_at, updated_at, archived_at
		FROM uicomponents 
		WHERE user_id = $1 AND archived_at IS NULL
		ORDER BY created_at DESC`

	rows, err := cs.db.Query(sqlQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query components by user: %w", err)
	}
	defer rows.Close()

	var components []UIComponent
	for rows.Next() {
		var component UIComponent
		err := rows.Scan(
			&component.ID,
			&component.Title,
			&component.Type,
			&component.Code,
			&component.IsPublic,
			&component.UserID,
			&component.CreatedAt,
			&component.UpdatedAt,
			&component.ArchivedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan component row: %w", err)
		}
		components = append(components, component)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over component rows: %w", err)
	}

	return components, nil
}

func (cs *UIComponentsStore) GetComponentsByUserPaginated(userID int, limit int, offset int) ([]UIComponent, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM uicomponents WHERE user_id = $1 AND archived_at IS NULL`
	var total int
	err := cs.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get paginated results
	sqlQuery := `
        SELECT id, title, type, code, user_id, created_at, updated_at
        FROM uicomponents 
        WHERE user_id = $1 AND archived_at IS NULL
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3`

	rows, err := cs.db.Query(sqlQuery, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query components: %w", err)
	}
	defer rows.Close()

	var components []UIComponent
	for rows.Next() {
		var component UIComponent
		err := rows.Scan(
			&component.ID, &component.Title, &component.Type, &component.Code,
			&component.UserID, &component.CreatedAt, &component.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan component: %w", err)
		}
		components = append(components, component)
	}

	return components, total, rows.Err()
}

// GetAllComponents retrieves all non-archived UI components (admin function)
func (cs *UIComponentsStore) GetAllComponents() ([]UIComponent, error) {
	sqlQuery := `
		SELECT id, title, type, code, is_public, user_id, created_at, updated_at, archived_at
		FROM uicomponents 
		WHERE archived_at IS NULL
		ORDER BY created_at DESC`

	rows, err := cs.db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query all components: %w", err)
	}
	defer rows.Close()

	var components []UIComponent
	for rows.Next() {
		var component UIComponent
		err := rows.Scan(
			&component.ID,
			&component.Title,
			&component.Type,
			&component.Code,
			&component.IsPublic,
			&component.UserID,
			&component.CreatedAt,
			&component.UpdatedAt,
			&component.ArchivedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan component row: %w", err)
		}
		components = append(components, component)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over component rows: %w", err)
	}

	return components, nil
}

// GetAllPublicComponentsWithUser retrieves all public components with user first and last name
func (cs *UIComponentsStore) GetAllPublicComponentsWithUser() ([]PublicComponentWithUser, error) {
	sqlQuery := `
		SELECT 
			c.id, c.title, c.type, c.code, c.is_public, c.user_id,
			u.first_name, u.last_name,
			c.created_at, c.updated_at, c.archived_at
		FROM uicomponents c
		JOIN users u ON c.user_id = u.id
		WHERE c.is_public = TRUE AND c.archived_at IS NULL
		ORDER BY c.created_at DESC
	`

	rows, err := cs.db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query public components with user: %w", err)
	}
	defer rows.Close()

	var results []PublicComponentWithUser
	for rows.Next() {
		var pcwu PublicComponentWithUser
		err := rows.Scan(
			&pcwu.ID,
			&pcwu.Title,
			&pcwu.Type,
			&pcwu.Code,
			&pcwu.IsPublic,
			&pcwu.UserID,
			&pcwu.FirstName,
			&pcwu.LastName,
			&pcwu.CreatedAt,
			&pcwu.UpdatedAt,
			&pcwu.ArchivedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan public component row: %w", err)
		}
		results = append(results, pcwu)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over public component rows: %w", err)
	}

	return results, nil
}
