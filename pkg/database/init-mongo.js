db.createUser({
    user: "ayocodedb",
    pwd: "secret",
    roles: [
        {
            role: "readWrite",
            db: "acourse_course",
        }
    ],
    mechanisms: ["<SCRAM-SHA-1|SCRAM-SHA-256>"],
})