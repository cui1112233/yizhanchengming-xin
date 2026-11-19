/* REA ev_779279c5817eeb88a2d67e3f57474443d8120234a8bf6a6d26c678be05d3e56d; original SHA b5b22086...; constant comments are analyst annotations. */

longlong FUN_1801c0cf0(void)

{
  longlong *plVar1;
  longlong lVar2;
  
/* novels.prompts.inkos.reviser_prompts#127="dataclass" */
  plVar1 = (longlong *)FUN_1800015a0(DAT_1803ca608,DAT_1803bc208);
  if ((plVar1 == (longlong *)0x0) || (lVar2 = *plVar1, lVar2 == 0)) {
/* novels.prompts.inkos.reviser_prompts#127="dataclass" */
    plVar1 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803bc208);
    if (plVar1 == (longlong *)0x0) {
      return 0;
    }
    lVar2 = *plVar1;
  }
  return lVar2;
}
