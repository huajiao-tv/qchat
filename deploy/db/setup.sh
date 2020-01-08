 #!/bin/bash
 mongo << EOF
 use admin;
 db.createUser({ user: '${USER}', pwd: '${PASSWORD}', roles: [ { role: "userAdminAnyDatabase", db: "admin" } ] });

