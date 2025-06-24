ALTER TABLE uicomponents 
ADD CONSTRAINT fk_uicomponents_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) 
ON DELETE CASCADE ON UPDATE CASCADE;