// 0001_init.js – создаём базу, коллекцию и стартовые документы
db = db.getSiblingDB('kent_afcg');   // переключаемся на нужную БД

// Если коллекция users ещё не существует – создаём её
if (!db.getCollectionNames().includes('users')) {
    db.createCollection('users');
    print('Collection "users" created');
}

// Вставляем начальные документы
db.users.insertMany([
    { _id: 1, name: "admin", role: "superuser", createdAt: new Date() },
    { _id: 2, name: "guest", role: "readOnly", createdAt: new Date() }
]);
print('Initial users inserted');