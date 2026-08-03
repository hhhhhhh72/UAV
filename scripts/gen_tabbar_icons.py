import os
from PIL import Image, ImageDraw

SIZE = 81
OUT = r"D:\w-yao\miniprogram\static\tabbar"

def make_png(name, active=False):
    img = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    c = (26, 26, 26, 255) if active else (150, 151, 153, 255)  # #1a1a1a / #969799
    w = 4  # stroke width

    if name == "home":
        # House: roof triangle + body rectangle
        draw.polygon([(8,26), (40,8), (72,26)], outline=c, width=w)
        draw.rectangle([(18,26), (62,66)], outline=c, width=w)
        draw.line([(38,46), (38,66)], fill=c, width=w)

    elif name == "mall":
        # Shopping bag
        draw.rectangle([(14,16), (66,68)], outline=c, width=w)
        draw.arc([(26,4), (54,24)], 180, 0, fill=c, width=w)
        # Handle line
        draw.line([(26,14), (28,16)], fill=c, width=w)
        draw.line([(52,14), (54,16)], fill=c, width=w)

    elif name == "publish":
        # Plus in circle
        draw.ellipse([(10,10), (70,70)], outline=c, width=w)
        draw.line([(40,24), (40,56)], fill=c, width=w)
        draw.line([(24,40), (56,40)], fill=c, width=w)

    elif name == "shop":
        # Store: roof + door
        draw.polygon([(8,28), (40,10), (72,28)], outline=c, width=w)
        draw.rectangle([(14,28), (66,70)], outline=c, width=w)
        draw.rectangle([(30,38), (50,70)], fill=c)

    elif name == "mine":
        # Person: head circle + body arc
        draw.ellipse([(30,8), (50,28)], outline=c, width=w)
        draw.arc([(10,30), (70,80)], 180, 0, fill=c, width=w)

    suffix = "-active" if active else ""
    path = os.path.join(OUT, f"{name}{suffix}.png")
    img.save(path, "PNG")
    print(f"{name}{suffix}.png  {os.path.getsize(path)}B")

for name in ["home", "mall", "publish", "shop", "mine"]:
    make_png(name, active=False)
    if name == "publish":
        continue  # publish uses same for active/inactive
    make_png(name, active=True)

print("DONE")
