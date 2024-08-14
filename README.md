# Instructions

## Laptop Libp2p node

To run a libp2p server:

```
cd laptop
go run main.go
```

## Android libp2p node

1. Install `fyne` CLI from [here](https://github.com/fyne-io/fyne)

2. Move inside the `mobile` dir and package and install the apk

```
cd mobile

# Following generates an APK file
fyne package -os android

# Following instals the APK on your device
fyne install -os android
```

