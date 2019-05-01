# Frames to Video**
```
ffmpeg -i "<frame-folder>\frame_%015d.png" -c:v libx264 -vf fps=25 -pix_fmt yuv420p result.mp4
ffmpeg -i "<frame-folder>\frame_%015d.png" -r 25 -vcodec libx264 -profile:v high444 -refs 5 -crf 0 result.mp4
```

# Video to Frames**
```
ffmpeg -i result.mp4 "frame-folder\frame_%015d.png"
```