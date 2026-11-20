
longlong FUN_1800e3e20(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  int *piVar2;
  longlong *plVar3;
  undefined8 *puVar4;
  longlong *plVar5;
  longlong *plVar6;
  longlong lVar7;
  code *pcVar8;
  code *pcVar9;
  int iVar10;
  longlong lVar11;
  longlong local_res18 [2];

  plVar3 = DAT_1803b3e98;
  plVar1 = (longlong *)*param_3;
  local_res18[0] = 0;
  if (DAT_1803b3e98 == (longlong *)0x0) {
LAB_1800e3e74:
    DAT_1803b3e98 = (longlong *)FUN_18036cd10(param_1,DAT_1803b3910,DAT_1803ca778,8);
  }
  else if ((1 < *DAT_1803b3e98) || (DAT_1803b3e98[2] != 0)) {
    if ((DAT_1803b3e98 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b3e98 && (*DAT_1803b3e98 = *DAT_1803b3e98 + -1, *plVar3 == 0)))) {
      (**(code **)(plVar3[1] + 0x30))(plVar3);
    }
    goto LAB_1800e3e74;
  }
  plVar3 = DAT_1803b3e98;
  FUN_180001d00(param_1,DAT_1803b3e98);
  pcVar9 = (code *)0x0;
  lVar7 = *(longlong *)(DAT_1803ca780 + 0x20);
  if (*(char *)(lVar7 + 10) == '\0') {
/* Analyst constant binding: novels.pipeline.pipeline_runner#9="AgentSettingsStore" bytes=[3963940, 3963960] */
    puVar4 = (undefined8 *)FUN_1800015a0(DAT_1803ca780,DAT_1803b3f58);
    pcVar8 = pcVar9;
    if (puVar4 != (undefined8 *)0x0) {
      pcVar8 = (code *)*puVar4;
    }
LAB_1800e3f7d:
    if (pcVar8 == (code *)0x0) goto LAB_1800e3f82;
LAB_1800e3fc0:
    plVar5 = (longlong *)FUN_18037f0c0();
    if (plVar5 != (longlong *)0x0) {
      plVar6 = (longlong *)FUN_18037f0c0();
      if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
        (**(code **)(plVar5[1] + 0x30))(plVar5);
      }
      if (plVar6 != (longlong *)0x0) {
        *(undefined4 *)(plVar3 + 5) = 0x68;
        lVar7 = FUN_180376770(param_1,pcVar8,plVar6);
        if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
          (**(code **)(plVar6[1] + 0x30))(plVar6);
        }
        if (lVar7 != 0) {
          FUN_180001d80(param_1);
          if ((int)*plVar1 < 0) {
            return lVar7;
          }
          *plVar1 = *plVar1 + -1;
          if (*plVar1 != 0) {
            return lVar7;
          }
          (**(code **)(plVar1[1] + 0x30))(plVar1);
          return lVar7;
        }
      }
    }
    lVar7 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    iVar10 = *(int *)(lVar7 + 0xc);
    if (iVar10 == 0) {
      iVar10 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
      *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar10 + 1;
      *(int *)(lVar7 + 0xc) = iVar10;
    }
    if (iVar10 != DAT_1803a9868) {
      DAT_1803a9868 = iVar10;
/* Analyst constant binding: novels.pipeline.pipeline_runner#9="AgentSettingsStore" bytes=[3963940, 3963960] */
      DAT_1803b3c80 = FUN_1803707d0(lVar7,DAT_1803b3f58,*(undefined8 *)(DAT_1803b3f58 + 0x18));
    }
    if (-1 < DAT_1803b3c80) {
      lVar11 = (1L << (*(byte *)(lVar7 + 9) & 0x3f)) + lVar7;
      pcVar8 = *(code **)(lVar11 + 0x28 + DAT_1803b3c80 * 0x10);
      if (pcVar8 != (code *)0x0) goto LAB_1800e3fc0;
/* Analyst constant binding: novels.pipeline.pipeline_runner#9="AgentSettingsStore" bytes=[3963940, 3963960] */
      DAT_1803b3c80 = FUN_1803707d0(lVar7,DAT_1803b3f58,*(undefined8 *)(DAT_1803b3f58 + 0x18));
      if (-1 < DAT_1803b3c80) {
        pcVar8 = *(code **)(lVar11 + 0x28 + DAT_1803b3c80 * 0x10);
        goto LAB_1800e3f7d;
      }
    }
LAB_1800e3f82:
/* Analyst constant binding: novels.pipeline.pipeline_runner#9="AgentSettingsStore" bytes=[3963940, 3963960] */
    puVar4 = (undefined8 *)FUN_1800015a0(DAT_1803ab778,DAT_1803b3f58);
    if ((puVar4 != (undefined8 *)0x0) && (pcVar8 = (code *)*puVar4, pcVar8 != (code *)0x0))
    goto LAB_1800e3fc0;
/* Analyst constant binding: novels.pipeline.pipeline_runner#9="AgentSettingsStore" bytes=[3963940, 3963960] */
    FUN_180373340(param_1,local_res18,DAT_1803b3f58);
    lVar7 = local_res18[0];
  }
  piVar2 = *(int **)(lVar7 + 0x28);
  if (piVar2 == (int *)0x0) {
    pcVar8 = (code *)FUN_18038e2a0(plVar3,0x68);
    plVar5 = *(longlong **)(lVar7 + 0x28);
    if (pcVar8 != _Py_NoneStruct_exref) {
LAB_1800e40a9:
      pcVar9 = pcVar8;
      pcVar8 = pcVar9;
      if ((pcVar9 != (code *)0x0) && (*(int *)pcVar9 + 1 != 0)) {
        *(int *)pcVar9 = *(int *)pcVar9 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar3) goto LAB_1800e40f2;
    pcVar8 = (code *)FUN_18038e2a0(plVar3,0x68);
    *(int **)(pcVar8 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar5 = *(longlong **)(lVar7 + 0x28);
    if (pcVar8 != _Py_NoneStruct_exref) goto LAB_1800e40a9;
  }
  *(code **)(lVar7 + 0x28) = pcVar9;
  if (((plVar5 != (longlong *)0x0) && (-1 < (int)*plVar5)) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)
     ) {
    (**(code **)(plVar5[1] + 0x30))();
  }
  if (((pcVar8 != (code *)0x0) && (-1 < *(int *)pcVar8)) &&
     (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1, *(longlong *)pcVar8 == 0)) {
    (**(code **)(*(longlong *)(pcVar8 + 8) + 0x30))(pcVar8);
  }
LAB_1800e40f2:
  FUN_18036d0a0(plVar3,&DAT_180394940,plVar1);
  plVar5 = DAT_1803b3e98;
  if (plVar3 == DAT_1803b3e98) {
    if ((-1 < (int)*DAT_1803b3e98) && (*DAT_1803b3e98 = *DAT_1803b3e98 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))();
    }
    DAT_1803b3e98 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar7;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
