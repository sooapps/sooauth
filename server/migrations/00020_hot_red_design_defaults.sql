-- +goose Up
UPDATE theme_config
SET accent_color = '#FF3B3B', updated_at = now()
WHERE lower(accent_color) IN ('#e8ff3f', '#7b72e9');

UPDATE tenant_themes
SET accent_color = '#FF3B3B', updated_at = now()
WHERE lower(accent_color) IN ('#e8ff3f', '#7b72e9');

-- +goose Down
UPDATE theme_config
SET accent_color = '#E8FF3F', updated_at = now()
WHERE lower(accent_color) = '#ff3b3b';

UPDATE tenant_themes
SET accent_color = '#E8FF3F', updated_at = now()
WHERE lower(accent_color) = '#ff3b3b';
