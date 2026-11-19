/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180298040 */

longlong * FUN_180298040(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  longlong *plVar3;
  longlong *plVar4;
  longlong lVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong *plVar8;
  code *pcVar9;
  code *pcVar10;
  int iVar11;
  int iVar12;
  longlong *plVar13;
  longlong lVar14;
  longlong lVar15;
  undefined8 uVar16;
  longlong *local_res18;
  longlong *local_res20;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  longlong *local_48;
  longlong *local_40;
  
  plVar6 = DAT_1803c3188;
  plVar1 = (longlong *)*param_3;
  plVar2 = (longlong *)param_3[1];
  plVar3 = (longlong *)param_3[2];
  plVar4 = (longlong *)param_3[3];
  local_res20 = (longlong *)0x0;
  local_res18 = DAT_180391e70;
  if (DAT_1803c3188 == (longlong *)0x0) {
LAB_1802980b1:
    DAT_1803c3188 = (longlong *)FUN_18036cd10(param_1,DAT_1803c3210,DAT_1803ca4b8,0x28);
  }
  else if ((1 < *DAT_1803c3188) || (DAT_1803c3188[2] != 0)) {
    if ((DAT_1803c3188 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803c3188 && (*DAT_1803c3188 = *DAT_1803c3188 + -1, *plVar6 == 0)))) {
      (**(code **)(plVar6[1] + 0x30))(plVar6);
    }
    goto LAB_1802980b1;
  }
  plVar6 = DAT_1803c3188;
  FUN_180001d00(param_1,DAT_1803c3188);
  lVar5 = *(longlong *)(DAT_1803ca4c0 + 0x20);
  if (*(char *)(lVar5 + 10) == '\0') {
    plVar8 = (longlong *)FUN_1800015a0();
    if (plVar8 == (longlong *)0x0) {
      lVar14 = 0;
    }
    else {
      lVar14 = *plVar8;
    }
LAB_1802981b6:
    if (lVar14 == 0) goto LAB_1802981bb;
LAB_180298204:
    uVar16 = 0x8a;
    *(undefined4 *)(plVar6 + 5) = 0x8a;
    local_68 = plVar1;
    local_60 = plVar2;
    local_58 = plVar3;
    local_50 = plVar4;
    local_58 = (longlong *)FUN_18037d650(param_1,lVar14,&local_68,&local_58,DAT_1803c34b8);
    if (local_58 != (longlong *)0x0) {
      local_res20 = local_58;
      local_68 = (longlong *)FUN_180292c10(param_1);
      if (local_68 == (longlong *)0x0) {
/* novels.services.runtime_state_store#36="save_runtime_state_snapshot" */
        FUN_180373340(param_1,&local_res18,DAT_1803c3400);
        uVar16 = 0x90;
        plVar8 = local_res18;
        goto LAB_180298307;
      }
      local_res18 = (longlong *)FUN_18037f0c0();
      uVar16 = 0x90;
      if (local_res18 != (longlong *)0x0) {
        *(undefined4 *)(plVar6 + 5) = 0x90;
        local_48 = plVar1;
        local_40 = local_res18;
        plVar8 = (longlong *)FUN_180377750(param_1,local_68,&local_48);
        if ((-1 < (int)*local_res18) && (*local_res18 = *local_res18 + -1, *local_res18 == 0)) {
          (**(code **)(local_res18[1] + 0x30))(local_res18);
        }
        if (plVar8 != (longlong *)0x0) {
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          FUN_180001d80(param_1);
          plVar6 = local_58;
          iVar11 = (int)*local_58 + 1;
          iVar12 = (int)*local_58;
          if (iVar11 != 0) {
            *(int *)local_58 = iVar11;
            iVar12 = iVar11;
          }
          if ((-1 < iVar12) && (*local_58 = *local_58 + -1, *local_58 == 0)) {
            (**(code **)(local_58[1] + 0x30))(local_58);
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
          if ((int)*plVar4 < 0) {
            return plVar6;
          }
          *plVar4 = *plVar4 + -1;
          if (*plVar4 != 0) {
            return plVar6;
          }
          (**(code **)(plVar4[1] + 0x30))(plVar4);
          return plVar6;
        }
      }
    }
    plVar8 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    iVar11 = *(int *)(lVar5 + 0xc);
    if (iVar11 == 0) {
      iVar11 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
      *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar11 + 1;
      *(int *)(lVar5 + 0xc) = iVar11;
    }
    if (iVar11 != DAT_1803a9e0c) {
      DAT_1803a9e0c = iVar11;
/* novels.services.runtime_state_store#58="build_runtime_state_artifacts" */
      DAT_1803c3178 = FUN_1803707d0(lVar5,DAT_1803c34b0,*(undefined8 *)(DAT_1803c34b0 + 0x18));
    }
    if (-1 < DAT_1803c3178) {
      lVar15 = (1L << (*(byte *)(lVar5 + 9) & 0x3f)) + lVar5;
      lVar14 = *(longlong *)(lVar15 + 0x28 + DAT_1803c3178 * 0x10);
      if (lVar14 != 0) goto LAB_180298204;
/* novels.services.runtime_state_store#58="build_runtime_state_artifacts" */
      DAT_1803c3178 = FUN_1803707d0(lVar5,DAT_1803c34b0,*(undefined8 *)(DAT_1803c34b0 + 0x18));
      if (-1 < DAT_1803c3178) {
        lVar14 = *(longlong *)(lVar15 + 0x28 + DAT_1803c3178 * 0x10);
        goto LAB_1802981b6;
      }
    }
LAB_1802981bb:
/* novels.services.runtime_state_store#58="build_runtime_state_artifacts" */
    plVar8 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803c34b0);
    if ((plVar8 != (longlong *)0x0) && (lVar14 = *plVar8, lVar14 != 0)) goto LAB_180298204;
/* novels.services.runtime_state_store#58="build_runtime_state_artifacts" */
    FUN_180373340(param_1,&local_res18,DAT_1803c34b0);
    uVar16 = 0x8a;
    plVar8 = local_res18;
  }
LAB_180298307:
  local_res18 = (longlong *)plVar8[5];
  if (local_res18 == (longlong *)0x0) {
    pcVar9 = (code *)FUN_18038e2a0(plVar6,uVar16);
    plVar13 = (longlong *)plVar8[5];
    if (pcVar9 == _Py_NoneStruct_exref) {
      pcVar10 = (code *)0x0;
    }
    else {
      pcVar10 = pcVar9;
      if ((pcVar9 != (code *)0x0) && (*(int *)pcVar9 + 1 != 0)) {
        *(int *)pcVar9 = *(int *)pcVar9 + 1;
      }
    }
  }
  else {
    if ((longlong *)local_res18[3] == plVar6) goto LAB_1802983cb;
    pcVar9 = (code *)FUN_18038e2a0(plVar6,uVar16);
    *(longlong **)(pcVar9 + 0x10) = local_res18;
    iVar11 = (int)*local_res18 + 1;
    if (iVar11 != 0) {
      *(int *)local_res18 = iVar11;
    }
    plVar13 = (longlong *)plVar8[5];
    if (pcVar9 == _Py_NoneStruct_exref) {
      pcVar10 = (code *)0x0;
    }
    else {
      pcVar10 = pcVar9;
      if ((pcVar9 != (code *)0x0) && (*(int *)pcVar9 + 1 != 0)) {
        *(int *)pcVar9 = *(int *)pcVar9 + 1;
      }
    }
  }
  plVar8[5] = (longlong)pcVar10;
  if (((plVar13 != (longlong *)0x0) && (-1 < (int)*plVar13)) &&
     (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
    (**(code **)(plVar13[1] + 0x30))();
  }
  if (((pcVar9 != (code *)0x0) && (-1 < *(int *)pcVar9)) &&
     (*(longlong *)pcVar9 = *(longlong *)pcVar9 + -1, *(longlong *)pcVar9 == 0)) {
    (**(code **)(*(longlong *)(pcVar9 + 8) + 0x30))(pcVar9);
  }
LAB_1802983cb:
  plVar7 = local_res20;
  FUN_18036d0a0(plVar6,"ooooo",plVar1,plVar2,plVar3,plVar4,local_res20);
  plVar13 = DAT_1803c3188;
  if (plVar6 == DAT_1803c3188) {
    if ((-1 < (int)*DAT_1803c3188) && (*DAT_1803c3188 = *DAT_1803c3188 + -1, *plVar13 == 0)) {
      (**(code **)(plVar13[1] + 0x30))();
    }
    DAT_1803c3188 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
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
  if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
    (**(code **)(plVar4[1] + 0x30))(plVar4);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar8;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (longlong *)0x0;
}
