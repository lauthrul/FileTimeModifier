# Description
Batch change file creation/modification/access time.

#Usage:
```
  -d string
        specifiy directory to process
  -f string
        specifiy file to process
  -r    recursive directory or not
  -s string
        time values to be set. 
            @C : creation time
            @M : modification time
            @A : access time
            @@ : auto choose earliest time from creation/modification/access time
        for example:
            @M=@C : change modification to creation time
            @C="2020/04/10 20:00:00" : change creation time to 2020/04/10 20:00:00
        multi-filed split with ",", for example:
            @C=@@,@M=@@ : change creation and modification time to the earliest one
```

#For Example:
Recursive directory `D:\Users\lauthrul\Pictures\Wallpaper` and change creation/modification/access time to the earliest one:
<br>
`FileTimeModifier.exe -d "D:\Users\lauthrul\Pictures\Wallpaper" -r -s "@C=@@,@M=@@,@A=@@"`

Change file creation time to the earliest one under directory `D:\Users\lauthrul\Pictures\Wallpaper` (no recursive):
<br>
`FileTimeModifier.exe -d "D:\Users\lauthrul\Pictures\Wallpaper" -s "@C="2020/04/10 20:00:00""`

Change file creation time to the earliest one:
<br>
`FileTimeModifier.exe -f "D:\Users\lauthrul\Pictures\Wallpaper\test.jpg" -s "@C=@@"`