from pathlib import Path
source=Path('/out/town2-formal-gui.py').read_text()
for kind in ['town','world']:
    code=source.replace("'formal-gui-town2'",repr('formal-gui-'+kind))
    code=code.replace("KOL_POSITION_CITY_GUI='1'",'KOL_MAP_GUI_KIND='+repr(kind))
    code=code.replace("'bin/town2-gallery'","'bin/maps-gui'")
    code=code.replace('^TestHDPositionCityGUI$','^TestHDFullMapGUI$')
    code=code.replace("case='second-city'",'case='+repr('full-map-'+kind))
    exec(compile(code,'controlled-full-map-gui-'+kind,'exec'),{'__name__':'__main__'})
