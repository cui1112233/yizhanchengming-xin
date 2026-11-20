/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1802239f0 */

code * FUN_1802239f0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  longlong *plVar3;
  int *piVar4;
  longlong *plVar5;
  undefined8 uVar6;
  int iVar7;
  int iVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  longlong *plVar12;
  code *pcVar13;
  code *pcVar14;
  longlong lVar15;
  undefined8 uVar16;
  code *local_68;
  longlong *local_60;
  undefined8 local_48;
  longlong *local_40;

  plVar5 = DAT_1803bed68;
  plVar1 = (longlong *)*param_3;
  plVar2 = (longlong *)param_3[1];
  plVar3 = (longlong *)param_3[2];
  local_60 = (longlong *)0x0;
  local_68 = (code *)0x0;
  if (DAT_1803bed68 == (longlong *)0x0) {
LAB_180223a60:
    DAT_1803bed68 = (longlong *)FUN_18036cd10(param_1,DAT_1803bed48,DAT_1803ca5b0,0x28);
  }
  else if ((1 < *DAT_1803bed68) || (DAT_1803bed68[2] != 0)) {
    if ((DAT_1803bed68 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803bed68 && (*DAT_1803bed68 = *DAT_1803bed68 + -1, *plVar5 == 0)))) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
    goto LAB_180223a60;
  }
  plVar5 = DAT_1803bed68;
  FUN_180001d00(param_1,DAT_1803bed68);
  uVar16 = 0x117;
  *(undefined4 *)(plVar5 + 5) = 0x117;
/* novels.services.agent_settings#102="load" */
  plVar9 = (longlong *)FUN_18037e560(param_1,plVar1,DAT_1803bf1f0);
  if (plVar9 == (longlong *)0x0) {
    lVar15 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    goto LAB_180223f2f;
  }
  plVar10 = (longlong *)FUN_18037f0c0();
/* novels.services.agent_settings#5="modelRoutes" */
  uVar6 = DAT_1803beee8;
  local_60 = plVar9;
  if (plVar10 == (longlong *)0x0) {
    uVar16 = 0x118;
LAB_180223f23:
    lVar15 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    plVar11 = (longlong *)FUN_180370ec0(param_1);
    uVar16 = 0x118;
    local_48 = uVar6;
    *(undefined4 *)(plVar5 + 5) = 0x118;
    local_40 = plVar11;
    plVar12 = (longlong *)FUN_180377750(param_1,plVar10,&local_48);
    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))(plVar10);
    }
    if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))(plVar11);
    }
    if (plVar12 == (longlong *)0x0) goto LAB_180223f23;
    plVar10 = (longlong *)FUN_18037f0c0();
    if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
      (**(code **)(plVar12[1] + 0x30))(plVar12);
    }
    if (plVar10 == (longlong *)0x0) goto LAB_180223f23;
    *(undefined4 *)(plVar5 + 5) = 0x118;
    plVar11 = (longlong *)FUN_180376770(param_1,plVar10,plVar2);
    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))(plVar10);
    }
    if (plVar11 == (longlong *)0x0) goto LAB_180223f23;
    iVar7 = FUN_180001510(plVar11);
    if (iVar7 == -1) {
      lVar15 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        (**(code **)(plVar11[1] + 0x30))(plVar11);
      }
      goto LAB_180223f2f;
    }
    if (iVar7 != 1) {
      if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
        (**(code **)(plVar11[1] + 0x30))(plVar11);
      }
/* novels.services.agent_settings#8="" */
      plVar11 = DAT_1803bef00;
/* novels.services.agent_settings#8="" */
      iVar7 = (int)*DAT_1803bef00 + 1;
      if (iVar7 != 0) {
/* novels.services.agent_settings#8="" */
        *(int *)DAT_1803bef00 = iVar7;
      }
    }
    plVar10 = (longlong *)FUN_1803749a0(plVar11);
    if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))(plVar11);
    }
    if (plVar10 == (longlong *)0x0) goto LAB_180223f23;
    plVar11 = (longlong *)FUN_18037f0c0();
    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))(plVar10);
    }
    if (plVar11 == (longlong *)0x0) goto LAB_180223f23;
    *(undefined4 *)(plVar5 + 5) = 0x118;
    pcVar14 = (code *)FUN_180376210(param_1,plVar11);
    if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))(plVar11);
    }
    if (pcVar14 == (code *)0x0) goto LAB_180223f23;
    iVar7 = FUN_180001510(pcVar14);
    local_68 = pcVar14;
    if (iVar7 == -1) {
      uVar16 = 0x119;
      goto LAB_180223f23;
    }
    if (iVar7 != 0) {
      pcVar13 = pcVar14;
      if (*(int *)pcVar14 + 1 != 0) {
        *(int *)pcVar14 = *(int *)pcVar14 + 1;
        FUN_180001d80(param_1);
        goto LAB_18022410e;
      }
LAB_1802240ec:
      FUN_180001d80(param_1);
LAB_18022410e:
      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
        (**(code **)(plVar9[1] + 0x30))();
      }
      if ((-1 < *(int *)pcVar14) &&
         (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
        (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
      }
      if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
        (**(code **)(plVar1[1] + 0x30))();
      }
      if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
        (**(code **)(plVar2[1] + 0x30))(plVar2);
      }
      if ((int)*plVar3 < 0) {
        return pcVar13;
      }
      *plVar3 = *plVar3 + -1;
      if (*plVar3 != 0) {
        return pcVar13;
      }
      (**(code **)(plVar3[1] + 0x30))(plVar3);
      return pcVar13;
    }
    iVar7 = FUN_180001510(plVar3);
    if (iVar7 == -1) {
      uVar16 = 0x11b;
      goto LAB_180223f23;
    }
    iVar8 = 0;
    if (iVar7 == 0) {
LAB_180223e83:
      if (iVar8 == 0) {
        FUN_180001d80(param_1);
        pcVar13 = _Py_NoneStruct_exref;
        if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
          *(int *)_Py_NoneStruct_exref = *(int *)_Py_NoneStruct_exref + 1;
        }
        goto LAB_18022410e;
      }
      plVar10 = (longlong *)FUN_1803749a0(plVar3);
      if (plVar10 != (longlong *)0x0) {
        plVar11 = (longlong *)FUN_18037f0c0();
        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
          (**(code **)(plVar10[1] + 0x30))(plVar10);
        }
        uVar16 = 0x11c;
        if (plVar11 != (longlong *)0x0) {
          *(undefined4 *)(plVar5 + 5) = 0x11c;
          pcVar13 = (code *)FUN_180376210(param_1,plVar11);
          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
            (**(code **)(plVar11[1] + 0x30))(plVar11);
          }
          if (pcVar13 != (code *)0x0) goto LAB_1802240ec;
        }
        goto LAB_180223f23;
      }
      lVar15 = *(longlong *)(param_1 + 0x60);
      uVar16 = 0x11c;
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      plVar10 = (longlong *)FUN_1803749a0(plVar3);
      if (plVar10 == (longlong *)0x0) {
        lVar15 = *(longlong *)(param_1 + 0x60);
        uVar16 = 0x11b;
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
      else {
        plVar11 = (longlong *)FUN_18037f0c0();
        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
          (**(code **)(plVar10[1] + 0x30))(plVar10);
        }
        uVar16 = 0x11b;
        if (plVar11 == (longlong *)0x0) {
          lVar15 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
          *(undefined4 *)(plVar5 + 5) = 0x11b;
          plVar10 = (longlong *)FUN_180376210(param_1,plVar11);
          if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
            (**(code **)(plVar11[1] + 0x30))(plVar11);
          }
          if (plVar10 == (longlong *)0x0) goto LAB_180223f23;
          iVar8 = FUN_180001510(plVar10);
          if (iVar8 != -1) {
            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
              (**(code **)(plVar10[1] + 0x30))(plVar10);
            }
            goto LAB_180223e83;
          }
          lVar15 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
            (**(code **)(plVar10[1] + 0x30))(plVar10);
          }
        }
      }
    }
  }
LAB_180223f2f:
  piVar4 = *(int **)(lVar15 + 0x28);
  if (piVar4 == (int *)0x0) {
    pcVar14 = (code *)FUN_18038e2a0(plVar5,uVar16);
    plVar9 = *(longlong **)(lVar15 + 0x28);
    pcVar13 = pcVar14;
    if (pcVar14 == _Py_NoneStruct_exref) {
      pcVar13 = (code *)0x0;
    }
    else {
LAB_180223f89:
      pcVar14 = pcVar13;
      if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
        *(int *)pcVar13 = *(int *)pcVar13 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar4 + 6) == plVar5) goto LAB_180223fd2;
    pcVar14 = (code *)FUN_18038e2a0(plVar5,uVar16);
    *(int **)(pcVar14 + 0x10) = piVar4;
    if (*piVar4 + 1 != 0) {
      *piVar4 = *piVar4 + 1;
    }
    plVar9 = *(longlong **)(lVar15 + 0x28);
    pcVar13 = pcVar14;
    if (pcVar14 != _Py_NoneStruct_exref) goto LAB_180223f89;
    pcVar13 = (code *)0x0;
  }
  *(code **)(lVar15 + 0x28) = pcVar13;
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))();
  }
  if (((pcVar14 != (code *)0x0) && (-1 < *(int *)pcVar14)) &&
     (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
    (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
  }
LAB_180223fd2:
  FUN_18036d0a0(plVar5,"ooooo",plVar1,plVar2,plVar3,local_60,local_68);
  plVar9 = DAT_1803bed68;
  if (plVar5 == DAT_1803bed68) {
    if ((-1 < (int)*DAT_1803bed68) && (*DAT_1803bed68 = *DAT_1803bed68 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))();
    }
    DAT_1803bed68 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((local_60 != (longlong *)0x0) && (-1 < (int)*local_60)) &&
     (*local_60 = *local_60 + -1, *local_60 == 0)) {
    (**(code **)(local_60[1] + 0x30))(local_60);
  }
  if (((local_68 != (code *)0x0) && (-1 < *(int *)local_68)) &&
     (*(longlong *)local_68 = *(longlong *)local_68 + -1, *(longlong *)local_68 == 0)) {
    (**(code **)(*(longlong *)(local_68 + 8) + 0x30))(local_68);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  if ((-1 < (int)*plVar3) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)) {
    (**(code **)(plVar3[1] + 0x30))(plVar3);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar15;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (code *)0x0;
}
