DO $$ 
BEGIN
    CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS pull_requests (
    id VARCHAR(255) PRIMARY KEY,
    pull_requests_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) REFERENCES users(id) ON DELETE CASCADE,
    status pr_status NOT NULL DEFAULT 'OPEN',
    need_more_reviewers BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    merged_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pr_author ON pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_pr_status ON pull_requests(status);