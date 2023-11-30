# Matrix

Реализация библиотеки matrix.h.

## Содержание

0. [Введение](#введение)
1. [Реализация](#реализация)
2. [Информация](#информация)

![abbrevia](misc/images/Abbrevia.png)

## Введение

В данном проекте я реализовал умного телеграм-бота, способного хранить базу данных аббревиатур. Он быстро отвечает на запросы, предоставляя расшифровки аббревиатур. Кроме того, пользователи могут расширить базу данных, добавляя свои аббревиатуры с помощью команды /add. Благодаря этому боту, пользователи могут легко и быстро получить нужную информацию.

## Реализация

- Телеграм бот разработан на языке Golang стандарта C11 с использованием компилятора gcc.
- Хранение аббревиатур реализовано для БД PostgresSQL
- Для запуска со своим токеном бота и со своим БД, нужно вписать нужные данные в файле "credentials.txt"
- Предусмотрен Makefile для сборки или запуска проекта (с целями build, run).


## Информация 

Для того чтобы начать пользоваться ботом, просто напишите ему название аббревиатуры. Если аббревиатура имеется в базе данных, то бот выведет результат

Пример:

![abbrevia-gift(1)](misc/gifts/Abbrevia(1).gif)

Если результат не найден, бот предложит добавить новую аббревиатуру. Появляется клавиатура, где пользователь делает выбор добавлять не найденную аббревиатуру или нет. Если пользователь соглашается добавить новую аббревиатуру, он должен сделать это в формате [аббревиатура] [значение]

Пример:

![abbrevia](misc/gifts/Abbrevia(2).gif)

Чтобы добавить дополнительное обозначение аббревиатуры, нужно использовать команду /add [аббревиатура] [значение]. Если в базе данных уже присутствуют значения по аббревиатуре, которую хочет добавить пользователь, то бот сначала выведет имеющиеся значения по ней, а потом попросит подтвердить добавление. Нужно это для того, чтобы не добавлять схожие значения по аббревиатурам

Пример:

![abbrevia](misc/gifts/Abbrevia(3).gif)

Если при использовании команды /add по аббревиатуре введенной пользователем не найдено совпадений, то новая запись сразу же будет добавлена в базу данных

Пример:

![abbrevia](misc/gifts/Abbrevia(4).gif)