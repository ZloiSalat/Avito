# Avito
Для начала требуется склонировать данный репозиторий при помощи команды - git clone 
Далее запустить main.go файл (VS code, JB Goland или терминал) и вписать команду "docker-compose build -d"
Следующий шаг - запустить контейнер, командой "docker-compose up -d"
Полученно сообщение:

✔ Container avito-avito-db-1        Started                                                                                                                                                                                        
✔ Container avito-avito-test-app-1  Started  

 Это означает что вы молодец и можно приступать к проверке возможности программы.
 
 Дисклеймер! Все запросы, который будем передавать, начинаются с "localhost:8000". Во время тестирования, для совершения запросов я использовал Postman.
 
 В программе можно аппелировать 4 методами:

r.POST("/segment", CreateSegment)

r.POST("/manageUserSegments", AddUserToSegment)

r.GET("/segments/:id", GetActiveSegments)

r.DELETE("/segment/:slug", DeleteSegment)

 Начнем с самого первого:
1. r.POST("/segment", CreateSegment).
Данный метод, исходя из своего названия, создает в базе данных сегмент. При введении в URL строку "localhost:8000/segment"  в теле запроса требуется в значении ключа slug вписать желаемое название сегмента.

Пример: `{
    "slug": "Hello"
}`

В ответе получим `{
    "message": "Segment created successfully!"
}`

2. r.DELETE("/segment/:slug", DeleteSegment).
 
Данный метод удаляет сегмент из базы данных. Для того, чтобы метод сработал в URL запросе нужно написать следующее:
Пример: `localhost:8000/segment/Hello`
после чего мы должны получить сообщение: 

`{
    "message": "Segment deleted successfully!"
}`

и данного сегмента в базе данных, естественно, не будет.

3. r.POST("/manageUserSegments", AddUserToSegment).

Данный метод принимает список названий сегментов ,которые нужно добавить пользователю, список названий сегментов, которые нужно удалить у пользователя и id этого пользователя. 
Важно, таблице segments уже должны быть сегменты, которые мы хотим добавить в таблицу user_segments. В теле запроса пишем cледующее.

Пример: `{
    "user_id": 1000,
    "add_segments": ["Hello", "World"],
    "remove_segments": ["World"]
}`

В ответ должно быть: `{
    "message": "Ok!"
}`

4. r.GET("/segments/:id", GetActiveSegments).
В данном методе можно получить активные сегменты, которые есть у юзера.

Пример: `localhost:8000/segments/1000`

в ответе получим: `{
    "segments": [
        "Hello"
    ]
}`
