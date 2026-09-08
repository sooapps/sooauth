-- +goose Up
UPDATE theme_config
SET accent_color = '#E8FF3F', updated_at = now()
WHERE lower(accent_color) = '#7b72e9';

UPDATE tenant_themes
SET accent_color = '#E8FF3F', updated_at = now()
WHERE lower(accent_color) = '#7b72e9';

-- +goose Down
UPDATE theme_config
SET accent_color = '#7B72E9', updated_at = now()
WHERE lower(accent_color) = '#e8ff3f';

UPDATE tenant_themes
SET accent_color = '#7B72E9', updated_at = now()
WHERE lower(accent_color) = '#e8ff3f';
