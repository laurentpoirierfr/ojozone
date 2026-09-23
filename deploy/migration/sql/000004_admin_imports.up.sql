BEGIN;

-- Structure des imports administrés : l'alimentation des données se fait via
-- l'API (fichiers JSON + /admin/imports). Aucune donnée de référence n'est ici.

CREATE TABLE imports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type text NOT NULL,
    status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'validated', 'published')),
    created_by uuid REFERENCES users(id) ON DELETE SET NULL,
    line_count integer NOT NULL DEFAULT 0 CHECK (line_count >= 0),
    valid_count integer NOT NULL DEFAULT 0 CHECK (valid_count >= 0),
    invalid_count integer NOT NULL DEFAULT 0 CHECK (invalid_count >= 0),
    report jsonb NOT NULL DEFAULT '[]',
    created_at timestamptz NOT NULL DEFAULT now(),
    validated_at timestamptz,
    published_at timestamptz
);

CREATE TABLE import_rows (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    import_id uuid NOT NULL REFERENCES imports(id) ON DELETE CASCADE,
    line_number integer NOT NULL CHECK (line_number > 0),
    payload jsonb NOT NULL,
    valid boolean,
    error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (import_id, line_number)
);

COMMIT;