
longlong FUN_1801c0cf0(void)

{
  longlong *plVar1;
  longlong lVar2;
  
  plVar1 = (longlong *)FUN_1800015a0(DAT_1803ca608,DAT_1803bc208);
  if ((plVar1 == (longlong *)0x0) || (lVar2 = *plVar1, lVar2 == 0)) {
    plVar1 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803bc208);
    if (plVar1 == (longlong *)0x0) {
      return 0;
    }
    lVar2 = *plVar1;
  }
  return lVar2;
}

