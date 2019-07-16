# mystem-docker [![Go Report Card](https://goreportcard.com/badge/github.com/azzzak/mystem-docker)](https://goreportcard.com/report/github.com/azzzak/mystem-docker)

Программа [MyStem](https://yandex.ru/dev/mystem/) — морфологический анализатор русского текста от Яндекса. Задача mystem-docker: упаковать MyStem в Docker-контейнер и работать с программой по http-протоколу.

Перед использованием необходимо прочитать и принять [лицензионное соглашение](https://yandex.ru/legal/mystem/) MyStem.

## Настройки

| Параметр           |   Тип    | Значение по умолчанию | Комментарий                                                      |
| :----------------- | :------: | :-------------------- | ---------------------------------------------------------------- |
| USER_DICT          | _string_ | -                     | Подключить [пользовательский словарь](#пользовательский-словарь) |
| GLUE_GRAMMEMES     |  _bool_  | false                 | Объединить словоформы при одной лемме                            |
| HOMONYMS_DETECTION |  _bool_  | false                 | Применить контекстное снятие омонимии                            |
| TIMEOUT            |  _int_   | 1000                  | Ограничить время обработки каждого запроса (в миллисекундах)     |

Таймаут необходим в силу однопоточной работы приложения, чтобы не допустить его зависания при ошибке.

## Примеры запуска

Установить таймаут на 800 миллисекунд:

`docker run -p 2345:8080 -e TIMEOUT=800 azzzak/mystem`

Подключить словарь:

`docker run -v ~/dict:/stem/dict -p 2345:8080 -e USER_DICT=dict.txt azzzak/mystem`

Пример запуска в проде:

`docker run -d --restart always -v ~/dict:/stem/dict -p 127.0.0.1:2345:8080 -e USER_DICT=dict.txt -e HOMONYMS_DETECTION=true -e GLUE_GRAMMEMES=true -e TIMEOUT=800 --name mystem azzzak/mystem`

## Проверка

`curl -i -d "text=съешь еще этих мягких французских булок" -X POST http://localhost:2345/mystem`

## Использование

Для получения морфологического анализа надо отправить `POST` запрос с полем `text=[текст для анализа]` на `/mystem`. Ответ приходит в формате json. Об ошибке сигнализирует ответ со статус-кодом, отличным от `200`.

## JSON Schema ответа

```
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "MyStem JSON Schema",
  "type": "array",
  "items": {
    "type": "object",
    "properties": {
      "analysis": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "lex": {
              "description": "лемма",
              "type": "string"
            },
            "wt": {
              "description": "бесконтекстная вероятность леммы",
              "type": "number",
              "minimum": 0,
              "maximum": 1
            },
            "qual": {
              "description": "особые отметки",
              "type": "string"
            },
            "gr": {
              "description": "граммемы",
              "type": "string"
            }
          },
          "required": [
            "lex", "wt", "gr"
          ]
        },
        "text": {
          "description": "исходная словоформа",
          "type": "string"
        }
      }
    },
    "required": [
      "analysis", "text"
    ]
  }
}
```

[Расшифровка](https://yandex.ru/dev/mystem/doc/grammemes-values-docpage/) обозначений граммем.

## Пользовательский словарь

В случае некорректной работы с неологизмами словарь можно дополнить. Формат пользовательского словаря описан [в документации](https://yandex.ru/dev/mystem/doc/usage-examples-docpage/#usage-examples__dicts) MyStem.

**После изменения словаря нужно перезапустить контейнер.**
