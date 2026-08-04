BEGIN;

INSERT INTO users(id,external_id,display_name,email,enabled) VALUES
('00000000-0000-0000-0000-000000000001','alice','Alice Example','alice@example.test',true),
('00000000-0000-0000-0000-000000000002','bob','Bob Example','bob@example.test',true),
('00000000-0000-0000-0000-000000000003','carol','Carol Example','carol@example.test',true),
('00000000-0000-0000-0000-000000000004','disabled-user','Disabled User',NULL,false)
ON CONFLICT(id) DO UPDATE SET display_name=EXCLUDED.display_name,email=EXCLUDED.email,enabled=EXCLUDED.enabled,updated_at=now();

INSERT INTO groups(id,external_id,display_name,enabled) VALUES
('00000000-0000-0000-0000-000000000101','engineering','Engineering',true),
('00000000-0000-0000-0000-000000000102','support','Support',true),
('00000000-0000-0000-0000-000000000103','management','Management',true)
ON CONFLICT(id) DO UPDATE SET display_name=EXCLUDED.display_name,enabled=EXCLUDED.enabled,updated_at=now();

INSERT INTO group_memberships(user_id,group_id) VALUES
('00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000101'),
('00000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000102'),
('00000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000103')
ON CONFLICT DO NOTHING;

INSERT INTO documents(id,source_type,source_id,title,content_hash,status,is_public) VALUES
('00000000-0000-0000-0000-000000001001','fixture','public-release-notes','Public release notes','sha256:public','active',true),
('00000000-0000-0000-0000-000000001002','fixture','restricted-incident','Engineering incident report','sha256:incident','active',true),
('00000000-0000-0000-0000-000000001003','fixture','support-runbook','Support runbook','sha256:support','active',false),
('00000000-0000-0000-0000-000000001004','fixture','disabled-document','Disabled document','sha256:disabled','disabled',true),
('00000000-0000-0000-0000-000000001005','fixture','deleted-document','Deleted document','sha256:deleted','deleted',true)
ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title,status=EXCLUDED.status,is_public=EXCLUDED.is_public,updated_at=now();

INSERT INTO document_acl_users(document_id,user_id,decision) VALUES
('00000000-0000-0000-0000-000000001002','00000000-0000-0000-0000-000000000001','deny')
ON CONFLICT(document_id,user_id) DO UPDATE SET decision=EXCLUDED.decision;

INSERT INTO document_acl_groups(document_id,group_id,decision) VALUES
('00000000-0000-0000-0000-000000001003','00000000-0000-0000-0000-000000000102','allow'),
('00000000-0000-0000-0000-000000001002','00000000-0000-0000-0000-000000000101','allow')
ON CONFLICT(document_id,group_id) DO UPDATE SET decision=EXCLUDED.decision;

COMMIT;
