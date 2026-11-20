
longlong FUN_1802231f0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  int *piVar2;
  longlong *plVar3;
  int iVar4;
  longlong lVar5;
  longlong *plVar6;
  longlong *plVar7;
  code *pcVar8;
  code *pcVar9;
  code *pcVar10;
  undefined8 uVar11;
  longlong local_res18 [2];
  longlong *local_48;
  code *local_40;

  plVar3 = DAT_1803bed98;
  pcVar8 = (code *)0x0;
  plVar1 = (longlong *)*param_3;
  local_res18[0] = 0;
  if (DAT_1803bed98 == (longlong *)0x0) {
LAB_180223248:
    DAT_1803bed98 = (longlong *)FUN_18036cd10(param_1,DAT_1803bedc0,DAT_1803ca5b0,0x10);
  }
  else if ((1 < *DAT_1803bed98) || (DAT_1803bed98[2] != 0)) {
    if ((DAT_1803bed98 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803bed98 && (*DAT_1803bed98 = *DAT_1803bed98 + -1, *plVar3 == 0)))) {
      (**(code **)(plVar3[1] + 0x30))(plVar3);
    }
    goto LAB_180223248;
  }
  plVar3 = DAT_1803bed98;
  FUN_180001d00(param_1,DAT_1803bed98);
  lVar5 = FUN_180219800(param_1);
  if (lVar5 == 0) {
/* Analyst constant binding: novels.services.agent_settings#98="FileRepository" bytes=[4319419, 4319435] */
    FUN_180373340(param_1,local_res18,DAT_1803bf1d0);
    uVar11 = 0x10e;
    lVar5 = local_res18[0];
  }
  else {
    plVar6 = (longlong *)FUN_18037f0c0();
    if (plVar6 == (longlong *)0x0) {
      lVar5 = *(longlong *)(param_1 + 0x60);
      uVar11 = 0x10e;
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      plVar7 = (longlong *)FUN_18037f0c0();
      if (plVar7 == (longlong *)0x0) {
        lVar5 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
          (**(code **)(plVar6[1] + 0x30))(plVar6);
        }
        uVar11 = 0x10e;
      }
      else {
        uVar11 = 0x10e;
        local_40 = _Py_NoneStruct_exref;
        *(undefined4 *)(plVar3 + 5) = 0x10e;
        local_48 = plVar7;
        pcVar8 = (code *)FUN_180377750(param_1,plVar6,&local_48);
        if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
          (**(code **)(plVar6[1] + 0x30))(plVar6);
        }
        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))(plVar7);
        }
        if (pcVar8 == (code *)0x0) {
          lVar5 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          pcVar8 = (code *)0x0;
        }
        else {
          lVar5 = FUN_180219f00(param_1);
          if (lVar5 == 0) {
/* Analyst constant binding: novels.services.agent_settings#100="merge_agent_settings" bytes=[4319446, 4319468] */
            FUN_180373340(param_1,local_res18,DAT_1803bf1e0);
            uVar11 = 0x10f;
            lVar5 = local_res18[0];
          }
          else {
            iVar4 = PyObject_IsInstance(pcVar8,PyDict_Type_exref);
            pcVar9 = _Py_NoneStruct_exref;
            uVar11 = 0x10f;
            if (iVar4 != -1) {
              *(undefined4 *)(plVar3 + 5) = 0x10f;
              if (iVar4 != 0) {
                pcVar9 = pcVar8;
              }
              lVar5 = FUN_180376770(param_1,lVar5,pcVar9);
              if (lVar5 != 0) {
                FUN_180001d80(param_1);
                if ((-1 < *(int *)pcVar8) &&
                   (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1, *(longlong *)pcVar8 == 0)) {
                  (**(code **)(*(longlong *)(pcVar8 + 8) + 0x30))(pcVar8);
                }
                if ((int)*plVar1 < 0) {
                  return lVar5;
                }
                *plVar1 = *plVar1 + -1;
                if (*plVar1 != 0) {
                  return lVar5;
                }
                (**(code **)(plVar1[1] + 0x30))(plVar1);
                return lVar5;
              }
            }
            lVar5 = *(longlong *)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
          }
        }
      }
    }
  }
  piVar2 = *(int **)(lVar5 + 0x28);
  if (piVar2 == (int *)0x0) {
    pcVar9 = (code *)FUN_18038e2a0(plVar3,uVar11);
    plVar6 = *(longlong **)(lVar5 + 0x28);
    pcVar10 = pcVar9;
    if (pcVar9 == _Py_NoneStruct_exref) {
      pcVar10 = (code *)0x0;
    }
    else {
LAB_18022347f:
      pcVar9 = pcVar10;
      if ((pcVar10 != (code *)0x0) && (*(int *)pcVar10 + 1 != 0)) {
        *(int *)pcVar10 = *(int *)pcVar10 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar3) goto LAB_1802234c8;
    pcVar9 = (code *)FUN_18038e2a0(plVar3,uVar11);
    *(int **)(pcVar9 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar6 = *(longlong **)(lVar5 + 0x28);
    pcVar10 = pcVar9;
    if (pcVar9 != _Py_NoneStruct_exref) goto LAB_18022347f;
    pcVar10 = (code *)0x0;
  }
  *(code **)(lVar5 + 0x28) = pcVar10;
  if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)
     ) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if (((pcVar9 != (code *)0x0) && (-1 < *(int *)pcVar9)) &&
     (*(longlong *)pcVar9 = *(longlong *)pcVar9 + -1, *(longlong *)pcVar9 == 0)) {
    (**(code **)(*(longlong *)(pcVar9 + 8) + 0x30))(pcVar9);
  }
LAB_1802234c8:
  FUN_18036d0a0(plVar3,&DAT_180391be4,plVar1,pcVar8);
  plVar6 = DAT_1803bed98;
  if (plVar3 == DAT_1803bed98) {
    if ((-1 < (int)*DAT_1803bed98) && (*DAT_1803bed98 = *DAT_1803bed98 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))();
    }
    DAT_1803bed98 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((pcVar8 != (code *)0x0) && (-1 < *(int *)pcVar8)) &&
     (*(longlong *)pcVar8 = *(longlong *)pcVar8 + -1, *(longlong *)pcVar8 == 0)) {
    (**(code **)(*(longlong *)(pcVar8 + 8) + 0x30))(pcVar8);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar5;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return 0;
}
