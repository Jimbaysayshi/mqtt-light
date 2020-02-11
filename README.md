 
MQTTCLIENT

 +-----+---------+----------+---------+-----+
  | BCM |   Name  | Physical | Name    | BCM |
  +-----+---------+----++----+---------+-----+
  |     |    3.3v |  1 || 2  | 5v      |     |
  |   2 |   SDA 1 |  3 || 4  | 5v      |     |
  |   3 |   SCL 1 |  5 || 6  | 0v      |     |
  |   4 | GPIO  7 |  7 || 8  | TxD     | 14  |
  |     |      0v |  9 || 10 | RxD     | 15  |
  |  17 | GPIO  0 | 11 || 12 | GPIO  1 | 18  |
  |  27 | GPIO  2 | 13 || 14 | 0v      |     |
  |  22 | GPIO  3 | 15 || 16 | GPIO  4 | 23  |
  |     |    3.3v | 17 || 18 | GPIO  5 | 24  |
  |  10 |    MOSI | 19 || 20 | 0v      |     |
  |   9 |    MISO | 21 || 22 | GPIO  6 | 25  |
  |  11 |    SCLK | 23 || 24 | CE0     | 8   |
  |     |      0v | 25 || 26 | CE1     | 7   |
  |   0 |   SDA 0 | 27 || 28 | SCL 0   | 1   |
  |   5 | GPIO 21 | 29 || 30 | 0v      |     |
  |   6 | GPIO 22 | 31 || 32 | GPIO 26 | 12  |
  |  13 | GPIO 23 | 33 || 34 | 0v      |     |
  |  19 | GPIO 24 | 35 || 36 | GPIO 27 | 16  |
  |  26 | GPIO 25 | 37 || 38 | GPIO 28 | 20  |
  |     |      0v | 39 || 40 | GPIO 29 | 21  |
  +-----+---------+----++----+---------+-----+

1. rpio.Pin(17)
2. rpio.Pin(27)
3. rpio.Pin(22)
4. rpio.Pin(5)
5. rpio.Pin(23)
6. rpio.Pin(6)
7. rpio.Pin(13)
8. rpio.Pin(19)
9. rpio.Pin(26)
10. rpio.Pin(4)

---CONFIGURATION---
-ClientId is the name of the client
-Topic is passed to broker to filter messages for each client
-ID is the client identifier 
-Set Pins according to how many pins are used in total(ie. 3 lights and 1 functional = 4 pins)
-Pin mapping uses BCM map NOT GPIO/Name map 
-Function pin (ie. Buzzer) has to be next free pin after lights. If lights are not
used, function pin has to be set as pin 17
-Waitdelay is the delay in seconds which the program uses to reset the client function
-FunctionOntime is the time in milliseconds for how long function pin is on
-FunctionDelay is the time in seconds for how long function waits for next rotation
-FunctionRotation is the amount of on/off switches the function does before it exits