# GoSpeech

GoSpeech is a wrapper under Yandex.SpeechKit voice to text engine. It proxifies requests to SpeechKit API, handling long-running requests. It could prevent request timeouts in services, integrated with SpeechKit, because GoSpeech sends callback after file downloaded successfully.

## Usage example:

```
curl http://127.0.0.1:8989/\?speaker\=zahar\&key\=YOU_API_KEY_HERE\&format\=mp3\&lang\=ru-RU\&text\=%D0%94%D0%B0%D0%BB%D0%B5%D0%BA%D0%BE-%D0%B4%D0%B0%D0%BB%D0%B5%D0%BA%D0%BE%20%D0%B7%D0%B0%20%D1%81%D0%BB%D0%BE%D0%B2%D0%B5%D1%81%D0%BD%D1%8B%D0%BC%D0%B8%20%D0%B3%D0%BE%D1%80%D0%B0%D0%BC%D0%B8\&callback\=http://127.0.0.1:3000/callback
```

After first time call it returns empty result. When synthezed file will be downloaded callback to `callback` parameters address will be send through POST request with parameter `filename`, which will be contain relative filename path to mp3 file with voice to text downloaded file (for ex. `cache/7b5d310a4165757f87ebc862e2b09e1d.mp3`). It could be used to download and stream this file.

## Notes

GoSpeech also could reduce calls to SpeechKit API, for ex. when SpeechKit using in VoxImplant broadcast calls.

PS. It just my first and fast project written in Go lang, please, doesn't use it in production without code reviews and rewriting all code from scratch :)
