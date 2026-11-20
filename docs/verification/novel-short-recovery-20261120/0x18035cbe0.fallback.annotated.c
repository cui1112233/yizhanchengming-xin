/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x18035cbe0 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

longlong FUN_18035cbe0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  uint uVar1;
  int *piVar2;
  int *piVar3;
  longlong *plVar4;
  code *pcVar5;
  longlong *plVar6;
  char cVar7;
  int iVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  longlong lVar12;
  code *pcVar13;
  code *pcVar14;
  longlong *plVar15;
  undefined8 uVar16;
  code *pcVar17;
  int iVar18;
  undefined8 *puVar19;
  longlong lVar20;
  longlong lVar21;
  longlong lVar22;
  undefined4 uVar23;
  longlong lVar24;
  longlong *plVar25;
  undefined1 auStack_158 [32];
  longlong *local_138;
  longlong *local_130;
  code *local_128;
  longlong *local_120;
  undefined8 local_118;
  longlong *local_110;
  undefined4 local_108;
  longlong local_f8;
  longlong *local_f0;
  code *local_e8;
  longlong *local_e0;
  longlong *local_d8;
  code *local_d0;
  longlong *local_c8;
  longlong *local_c0;
  undefined8 local_b8;
  longlong *local_b0;
  longlong *local_a8;
  longlong *local_a0;
  longlong *local_98;
  longlong *local_90;
  longlong *local_88;
  longlong *local_78;
  undefined8 local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  code *local_48;
  ulonglong local_40;

  pcVar13 = _Py_NoneStruct_exref;
  local_40 = DAT_1803a6040 ^ (ulonglong)auStack_158;
  local_88 = (longlong *)*param_3;
  plVar9 = (longlong *)param_3[2];
  local_a0 = (longlong *)param_3[1];
  local_98 = (longlong *)param_3[3];
  local_c8 = (longlong *)0x0;
  local_c0 = (longlong *)0x0;
  local_78 = (longlong *)0x0;
  local_d0 = _Py_NoneStruct_exref;
  if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
    *(int *)_Py_NoneStruct_exref = *(int *)_Py_NoneStruct_exref + 1;
  }
  plVar15 = DAT_1803c9af8;
  local_d8 = DAT_1803c9af8;
  local_f8 = param_1;
  local_a8 = plVar9;
  if (DAT_1803c9af8 == (longlong *)0x0) {
LAB_18035cc93:
    DAT_1803c9af8 = (longlong *)FUN_18036cd10(param_1,DAT_1803c9a88,DAT_1803ca340,0x40);
  }
  else if ((1 < *DAT_1803c9af8) || (DAT_1803c9af8[2] != 0)) {
    if ((DAT_1803c9af8 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803c9af8 && (*DAT_1803c9af8 = *DAT_1803c9af8 + -1, *plVar15 == 0)))) {
      (**(code **)(plVar15[1] + 0x30))(plVar15);
    }
    goto LAB_18035cc93;
  }
  plVar15 = DAT_1803c9af8;
  local_d8 = DAT_1803c9af8;
  FUN_180001d00(param_1,DAT_1803c9af8);
  uVar16 = DAT_1803ab258;
  plVar9 = (longlong *)FUN_180380b80(plVar9,DAT_1803ab258);
  if (plVar9 == (longlong *)0x0) {
LAB_18035cce2:
    plVar15 = *(longlong **)(param_1 + 0x60);
    uVar23 = 0x48;
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    plVar10 = (longlong *)FUN_180375000(param_1,uVar16,plVar9,DAT_1803ab258);
    if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))(plVar9);
    }
    if (plVar10 == (longlong *)0x0) goto LAB_18035cce2;
    plVar9 = (longlong *)FUN_180001930();
    local_b0 = plVar9;
    if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
      (**(code **)(plVar10[1] + 0x30))(plVar10);
    }
    if (plVar9 == (longlong *)0x0) goto LAB_18035cce2;
    plVar11 = (longlong *)(**(code **)(plVar9[1] + 0xe0))(plVar9);
    plVar10 = (longlong *)0x0;
    local_90 = plVar11;
    if (plVar11 != (longlong *)0x0) {
      plVar25 = plVar15 + 5;
      local_108 = 0x4d;
      pcVar17 = pcVar13;
      local_f0 = plVar25;
      do {
        plVar4 = local_c8;
        plVar10 = local_f0;
        local_90 = plVar11;
        plVar6 = plVar11;
        if (((local_78 != (longlong *)0x0) && (plVar25 = plVar10, -1 < (int)*local_78)) &&
           (*local_78 = *local_78 + -1, *local_78 == 0)) {
          plVar25 = local_78 + 1;
          local_78 = plVar11;
          (**(code **)(*plVar25 + 0x30))();
          plVar25 = plVar10;
          plVar6 = local_78;
        }
        local_78 = plVar6;
        iVar8 = (int)*plVar11 + 1;
        if (iVar8 != 0) {
          *(int *)plVar11 = iVar8;
          plVar25 = plVar10;
        }
        local_c8 = plVar11;
        if (((plVar4 != (longlong *)0x0) && (plVar25 = plVar10, -1 < (int)*plVar4)) &&
           (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
          (**(code **)(plVar4[1] + 0x30))();
        }
        plVar10 = local_88;
        param_1 = local_f8;
        if (local_88 == (longlong *)0x0) {
          plVar9 = (longlong *)
                   PyUnicode_FromFormat
                             ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.utils.llm_retry#9="fn" */
                              ,DAT_1803c9b48);
          param_1 = local_f8;
          lVar12 = FUN_180376770(local_f8,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar9);
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          FUN_18000d1d0(param_1,lVar12);
        }
        else {
          *(undefined4 *)plVar25 = 0x4a;
          lVar12 = FUN_180376210(local_f8,local_88);
          if (lVar12 != 0) {
            if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
              (**(code **)(plVar11[1] + 0x30))(plVar11);
            }
            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
              (**(code **)(plVar9[1] + 0x30))(plVar9);
            }
            FUN_180001d80(param_1);
            if (((local_d0 != (code *)0x0) && (-1 < *(int *)local_d0)) &&
               (*(longlong *)local_d0 = *(longlong *)local_d0 + -1, *(longlong *)local_d0 == 0)) {
              (**(code **)(*(longlong *)(local_d0 + 8) + 0x30))(local_d0);
            }
            if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
              (**(code **)(plVar11[1] + 0x30))(plVar11);
            }
            if (((local_c0 != (longlong *)0x0) && (-1 < (int)*local_c0)) &&
               (*local_c0 = *local_c0 + -1, *local_c0 == 0)) {
              (**(code **)(local_c0[1] + 0x30))();
            }
            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
              (**(code **)(plVar10[1] + 0x30))(plVar10);
            }
            if ((-1 < (int)*local_a0) && (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
              (**(code **)(local_a0[1] + 0x30))();
            }
            if ((-1 < (int)*local_a8) && (*local_a8 = *local_a8 + -1, *local_a8 == 0)) {
              (**(code **)(local_a8[1] + 0x30))();
            }
            if ((int)*local_98 < 0) {
              return lVar12;
            }
            *local_98 = *local_98 + -1;
            if (*local_98 != 0) {
              return lVar12;
            }
            (**(code **)(local_98[1] + 0x30))();
            return lVar12;
          }
          lVar12 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        piVar3 = *(int **)(param_1 + 0xf8);
        local_e0 = (longlong *)0x0;
        if ((piVar3 != (int *)0x0) && (*piVar3 + 1 != 0)) {
          *piVar3 = *piVar3 + 1;
        }
        piVar2 = *(int **)(lVar12 + 0x28);
        pcVar13 = (code *)FUN_18038e2a0(plVar15,0x4a);
        if (piVar2 == (int *)0x0) {
          plVar9 = *(longlong **)(lVar12 + 0x28);
          if (pcVar13 == _Py_NoneStruct_exref) {
            pcVar14 = (code *)0x0;
          }
          else {
LAB_18035cee6:
            pcVar14 = pcVar13;
            if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
              *(int *)pcVar13 = *(int *)pcVar13 + 1;
            }
          }
        }
        else {
          *(int **)(pcVar13 + 0x10) = piVar2;
          if (*piVar2 + 1 != 0) {
            *piVar2 = *piVar2 + 1;
          }
          plVar9 = *(longlong **)(lVar12 + 0x28);
          if (pcVar13 != _Py_NoneStruct_exref) goto LAB_18035cee6;
          pcVar14 = (code *)0x0;
        }
        *(code **)(lVar12 + 0x28) = pcVar14;
        if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))();
        }
        if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
           (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
          (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
        }
        plVar9 = *(longlong **)(param_1 + 0xf8);
        *(longlong *)(param_1 + 0xf8) = lVar12;
        if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))();
        }
        lVar12 = *(longlong *)(*(longlong *)(param_1 + 0xf8) + 8);
        lVar22 = *(longlong *)PyExc_Exception_exref;
        uVar1 = *(uint *)(*(longlong *)(lVar22 + 8) + 0xa8);
        if ((uVar1 >> 0x1a & 1) != 0) {
          lVar21 = *(longlong *)(lVar22 + 0x10);
          lVar20 = 0;
          if (0 < lVar21) {
            plVar9 = (longlong *)(lVar22 + 0x18);
            lVar24 = lVar20;
            do {
              if ((-1 < *(int *)(*(longlong *)(*plVar9 + 8) + 0xa8)) ||
                 ((*(uint *)(*plVar9 + 0xa8) & 0x40000000) == 0)) goto LAB_18035d0b9;
              lVar24 = lVar24 + 1;
              plVar9 = plVar9 + 1;
            } while (lVar24 < lVar21);
          }
          if (0 < lVar21) {
            puVar19 = (undefined8 *)(lVar22 + 0x18);
            while (iVar8 = FUN_18000ce90(param_1,lVar12,*puVar19), iVar8 == 0) {
              lVar20 = lVar20 + 1;
              puVar19 = puVar19 + 1;
              if (lVar21 <= lVar20) goto LAB_18035d002;
            }
            goto LAB_18035d126;
          }
LAB_18035d002:
          cVar7 = FUN_180373290(param_1,&local_e0);
          local_108 = 0x49;
          if (cVar7 != '\0') {
            local_108 = 0;
          }
          lVar12 = local_e0[5];
          plVar15 = local_e0;
          if ((lVar12 != 0) && (*(longlong **)(lVar12 + 0x18) == local_d8)) {
            *(undefined4 *)(local_d8 + 5) = *(undefined4 *)(lVar12 + 0x24);
          }
LAB_18035da28:
          plVar9 = *(longlong **)(param_1 + 0xf8);
          *(int **)(param_1 + 0xf8) = piVar3;
          uVar23 = local_108;
          if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
             (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))();
            uVar23 = local_108;
          }
          goto LAB_18035dbbe;
        }
        if (((int)uVar1 < 0) && ((*(uint *)(lVar22 + 0xa8) & 0x40000000) != 0)) {
          if ((*(uint *)(*(longlong *)(lVar12 + 8) + 0xa8) & 0x40000000) != 0) {
            lVar12 = *(longlong *)(lVar12 + 8);
          }
          if (((*(int *)(*(longlong *)(lVar12 + 8) + 0xa8) < 0) &&
              ((*(uint *)(lVar12 + 0xa8) & 0x40000000) != 0)) &&
             ((*(uint *)(lVar22 + 0xa8) & 0x40000000) != 0)) {
            cVar7 = FUN_180382640(lVar12,lVar22);
          }
          else {
            cVar7 = lVar12 == lVar22;
          }
          if (cVar7 == '\0') goto LAB_18035d002;
        }
        else {
LAB_18035d0b9:
          pcVar13 = *(code **)PyExc_TypeError_exref;
          local_b8 = PyUnicode_FromString
                               (
                               "catching classes that do not inherit from BaseException is not allowed"
                               );
          plVar9 = *(longlong **)(param_1 + 0x60);
          local_e8 = pcVar13;
          if ((pcVar13 != _Py_NoneStruct_exref) && (pcVar13 != (code *)0x0)) {
            FUN_1803711b0(param_1,&local_e8,&local_b8,0);
          }
          *(undefined8 *)(param_1 + 0x60) = local_b8;
          if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
             (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
        }
LAB_18035d126:
        pcVar14 = local_d0;
        pcVar13 = *(code **)(param_1 + 0xf8);
        iVar8 = *(int *)pcVar13 + 1;
        iVar18 = *(int *)pcVar13;
        if (iVar8 != 0) {
          *(int *)pcVar13 = iVar8;
          iVar18 = iVar8;
        }
        if (iVar18 + 1 != 0) {
          *(int *)pcVar13 = iVar18 + 1;
        }
        pcVar5 = pcVar13;
        if (((local_d0 != (code *)0x0) && (-1 < *(int *)pcVar17)) &&
           (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0)) {
          pcVar17 = local_d0 + 8;
          local_d0 = pcVar13;
          (**(code **)(*(longlong *)pcVar17 + 0x30))(pcVar14);
          pcVar5 = local_d0;
        }
        local_d0 = pcVar5;
        plVar9 = local_90;
        if (local_a8 == (longlong *)0x0) {
          plVar9 = (longlong *)
                   PyUnicode_FromFormat
                             ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.utils.llm_retry#10="max_attempts" */
                              ,DAT_1803c9b50);
          plVar15 = (longlong *)
                    FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar9);
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          FUN_18000d1d0(param_1,plVar15);
LAB_18035da02:
          local_108 = 0x4d;
LAB_18035da0a:
          if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
             (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
            (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
          }
          goto LAB_18035da28;
        }
        iVar8 = FUN_180384e70(local_90,local_a8);
        if (iVar8 == -1) {
          plVar15 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_18035da02;
        }
        if (iVar8 == 1) {
LAB_18035d962:
          cVar7 = FUN_180373290(param_1,&local_e0);
          local_108 = 0x4e;
          if (cVar7 != '\0') {
            local_108 = 0;
          }
          lVar12 = local_e0[5];
          plVar15 = local_e0;
          if ((lVar12 != 0) && (*(longlong **)(lVar12 + 0x18) == local_d8)) {
            *(undefined4 *)(local_d8 + 5) = *(undefined4 *)(lVar12 + 0x24);
          }
          goto LAB_18035da0a;
        }
        lVar12 = *(longlong *)(DAT_1803ca338 + 0x20);
        if (*(char *)(lVar12 + 10) != '\0') {
          iVar8 = *(int *)(lVar12 + 0xc);
          if (iVar8 == 0) {
            iVar8 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
            *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar8 + 1;
            *(int *)(lVar12 + 0xc) = iVar8;
          }
          if (iVar8 != DAT_1803aa0b4) {
            DAT_1803aa0b4 = iVar8;
            DAT_1803c9aa0 =
/* novels.utils.llm_retry#11="is_transient_llm_error" */
                 FUN_1803707d0(lVar12,DAT_1803c9b58,*(undefined8 *)(DAT_1803c9b58 + 0x18));
          }
          if (-1 < DAT_1803c9aa0) {
            lVar21 = (1L << (*(byte *)(lVar12 + 9) & 0x3f)) + lVar12;
            lVar22 = *(longlong *)(lVar21 + 0x28 + DAT_1803c9aa0 * 0x10);
            if (lVar22 != 0) goto LAB_18035d291;
            DAT_1803c9aa0 =
/* novels.utils.llm_retry#11="is_transient_llm_error" */
                 FUN_1803707d0(lVar12,DAT_1803c9b58,*(undefined8 *)(DAT_1803c9b58 + 0x18));
            if (-1 < DAT_1803c9aa0) {
              lVar22 = *(longlong *)(lVar21 + 0x28 + DAT_1803c9aa0 * 0x10);
              goto LAB_18035d264;
            }
          }
LAB_18035d269:
/* novels.utils.llm_retry#11="is_transient_llm_error" */
          plVar15 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803c9b58);
          if ((plVar15 != (longlong *)0x0) && (lVar22 = *plVar15, lVar22 != 0)) goto LAB_18035d291;
/* novels.utils.llm_retry#11="is_transient_llm_error" */
          FUN_180373340(param_1,&local_e0,DAT_1803c9b58);
          plVar15 = local_e0;
          goto LAB_18035da0a;
        }
        plVar15 = (longlong *)FUN_1800015a0();
        if (plVar15 == (longlong *)0x0) {
          lVar22 = 0;
        }
        else {
          lVar22 = *plVar15;
        }
LAB_18035d264:
        if (lVar22 == 0) goto LAB_18035d269;
LAB_18035d291:
        param_1 = local_f8;
        plVar25 = local_d8 + 5;
        *(undefined4 *)plVar25 = 0x4d;
        pcVar17 = (code *)FUN_180376770(local_f8,lVar22,pcVar13);
        if (pcVar17 == (code *)0x0) {
          plVar15 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          local_108 = 0x4d;
          goto LAB_18035da0a;
        }
        if (pcVar17 == _Py_TrueStruct_exref) {
LAB_18035d2c7:
          iVar8 = 1;
        }
        else if ((pcVar17 == _Py_FalseStruct_exref) || (pcVar17 == _Py_NoneStruct_exref)) {
          iVar8 = 0;
        }
        else {
          lVar12 = *(longlong *)(pcVar17 + 8);
          if ((*(longlong *)(lVar12 + 0x60) == 0) ||
             (pcVar14 = *(code **)(*(longlong *)(lVar12 + 0x60) + 0x48), pcVar14 == (code *)0x0)) {
            if (((*(longlong **)(lVar12 + 0x70) == (longlong *)0x0) ||
                (pcVar14 = (code *)**(longlong **)(lVar12 + 0x70), pcVar14 == (code *)0x0)) &&
               ((*(longlong **)(lVar12 + 0x68) == (longlong *)0x0 ||
                (pcVar14 = (code *)**(longlong **)(lVar12 + 0x68), pcVar14 == (code *)0x0))))
            goto LAB_18035d2c7;
            lVar12 = (*pcVar14)(pcVar17);
          }
          else {
            iVar8 = (*pcVar14)(pcVar17);
            lVar12 = (longlong)iVar8;
          }
          if (lVar12 < 1) {
            iVar8 = -(uint)(lVar12 != 0);
          }
          else {
            iVar8 = 1;
          }
        }
        if ((-1 < *(int *)pcVar17) &&
           (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0)) {
          (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
        }
        if (iVar8 == -1) {
          plVar15 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_18035da02;
        }
        if (iVar8 == 0) goto LAB_18035d962;
        if (local_98 == (longlong *)0x0) {
          plVar9 = (longlong *)
                   PyUnicode_FromFormat
                             ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.utils.llm_retry#12="base_delay" */
                              ,DAT_1803c9b60);
          plVar15 = (longlong *)
                    FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar9);
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          FUN_18000d1d0(param_1,plVar15);
          local_108 = 0x4f;
          goto LAB_18035da0a;
        }
        plVar15 = (longlong *)FUN_180381a40(local_98,plVar9);
        plVar9 = local_c0;
        if (plVar15 == (longlong *)0x0) {
          plVar15 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          local_108 = 0x4f;
          goto LAB_18035da0a;
        }
        plVar10 = plVar15;
        if (((local_c0 != (longlong *)0x0) && (-1 < (int)*local_c0)) &&
           (*local_c0 = *local_c0 + -1, *local_c0 == 0)) {
          plVar10 = local_c0 + 1;
          local_c0 = plVar15;
          (**(code **)(*plVar10 + 0x30))(plVar9);
          plVar10 = local_c0;
        }
        local_c0 = plVar10;
        lVar12 = *(longlong *)(DAT_1803ca338 + 0x20);
        if (*(char *)(lVar12 + 10) == '\0') {
          plVar9 = (longlong *)FUN_1800015a0();
          if (plVar9 == (longlong *)0x0) {
            lVar12 = 0;
          }
          else {
            lVar12 = *plVar9;
          }
LAB_18035d476:
          if (lVar12 == 0) goto LAB_18035d47b;
        }
        else {
          iVar8 = *(int *)(lVar12 + 0xc);
          if (iVar8 == 0) {
            iVar8 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
            *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar8 + 1;
            *(int *)(lVar12 + 0xc) = iVar8;
          }
          if (iVar8 != DAT_1803aa0ac) {
            DAT_1803aa0ac = iVar8;
            DAT_1803c9ab0 =
/* novels.utils.llm_retry#13="logger" */
                 FUN_1803707d0(lVar12,DAT_1803c9b68,*(undefined8 *)(DAT_1803c9b68 + 0x18));
          }
          if (-1 < DAT_1803c9ab0) {
            lVar22 = (1L << (*(byte *)(lVar12 + 9) & 0x3f)) + lVar12;
            if (*(longlong *)(lVar22 + 0x28 + DAT_1803c9ab0 * 0x10) != 0) goto LAB_18035d4a3;
            DAT_1803c9ab0 =
/* novels.utils.llm_retry#13="logger" */
                 FUN_1803707d0(lVar12,DAT_1803c9b68,*(undefined8 *)(DAT_1803c9b68 + 0x18));
            if (-1 < DAT_1803c9ab0) {
              lVar12 = *(longlong *)(lVar22 + 0x28 + DAT_1803c9ab0 * 0x10);
              goto LAB_18035d476;
            }
          }
LAB_18035d47b:
/* novels.utils.llm_retry#13="logger" */
          plVar9 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803c9b68);
          if ((plVar9 == (longlong *)0x0) || (*plVar9 == 0)) {
/* novels.utils.llm_retry#13="logger" */
            FUN_180373340(param_1,&local_e0,DAT_1803c9b68);
            local_108 = 0x50;
            plVar15 = local_e0;
            goto LAB_18035da0a;
          }
        }
LAB_18035d4a3:
        plVar9 = (longlong *)FUN_18037f0c0();
        if (plVar9 == (longlong *)0x0) {
LAB_18035d8a1:
          plVar15 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          local_108 = 0x50;
          goto LAB_18035da0a;
        }
        if (local_a0 == (longlong *)0x0) {
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          plVar9 = (longlong *)
                   PyUnicode_FromFormat
                             ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.utils.llm_retry#16="label" */
                              ,DAT_1803c9b80);
          plVar15 = (longlong *)
                    FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar9);
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          FUN_18000d1d0(param_1,plVar15);
          local_108 = 0x52;
          goto LAB_18035da0a;
        }
/* novels.utils.llm_retry#15="[%s] transient LLM error (attempt %s/%s), retrying in %.1fs: %s" */
        local_70 = DAT_1803c9b78;
        *(undefined4 *)(local_d8 + 5) = 0x50;
        local_68 = local_a0;
        local_60 = local_90;
        local_58 = local_a8;
        local_50 = plVar15;
        local_48 = pcVar13;
        plVar15 = (longlong *)FUN_18037b2a0(param_1,plVar9,&local_70);
        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
        if (plVar15 == (longlong *)0x0) goto LAB_18035d8a1;
        if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
          (**(code **)(plVar15[1] + 0x30))(plVar15);
        }
        lVar12 = *(longlong *)(DAT_1803ca338 + 0x20);
        if (*(char *)(lVar12 + 10) == '\0') {
          plVar9 = (longlong *)FUN_1800015a0();
          if (plVar9 == (longlong *)0x0) {
            lVar22 = 0;
          }
          else {
            lVar22 = *plVar9;
          }
LAB_18035d60d:
          if (lVar22 == 0) goto LAB_18035d612;
        }
        else {
          iVar8 = *(int *)(lVar12 + 0xc);
          if (iVar8 == 0) {
            iVar8 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
            *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar8 + 1;
            *(int *)(lVar12 + 0xc) = iVar8;
          }
          if (iVar8 != DAT_1803aa0bc) {
            DAT_1803aa0bc = iVar8;
            DAT_1803c9ae0 =
/* novels.utils.llm_retry#17="time" */
                 FUN_1803707d0(lVar12,DAT_1803c9b88,*(undefined8 *)(DAT_1803c9b88 + 0x18));
          }
          if (-1 < DAT_1803c9ae0) {
            lVar21 = (1L << (*(byte *)(lVar12 + 9) & 0x3f)) + lVar12;
            lVar22 = *(longlong *)(lVar21 + 0x28 + DAT_1803c9ae0 * 0x10);
            if (lVar22 != 0) goto LAB_18035d63a;
            DAT_1803c9ae0 =
/* novels.utils.llm_retry#17="time" */
                 FUN_1803707d0(lVar12,DAT_1803c9b88,*(undefined8 *)(DAT_1803c9b88 + 0x18));
            if (-1 < DAT_1803c9ae0) {
              lVar22 = *(longlong *)(lVar21 + 0x28 + DAT_1803c9ae0 * 0x10);
              goto LAB_18035d60d;
            }
          }
LAB_18035d612:
/* novels.utils.llm_retry#17="time" */
          plVar9 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803c9b88);
          if ((plVar9 == (longlong *)0x0) || (lVar22 = *plVar9, lVar22 == 0)) {
/* novels.utils.llm_retry#17="time" */
            FUN_180373340(param_1,&local_e0,DAT_1803c9b88);
            local_108 = 0x54;
            plVar15 = local_e0;
            goto LAB_18035da0a;
          }
        }
LAB_18035d63a:
        *(undefined4 *)(local_d8 + 5) = 0x54;
/* novels.utils.llm_retry#18="sleep" */
        plVar9 = (longlong *)FUN_18037e6c0(param_1,lVar22,DAT_1803c9b90);
        if (plVar9 == (longlong *)0x0) {
          plVar15 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          local_108 = 0x54;
          goto LAB_18035da0a;
        }
        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
        if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
           (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
          (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
        }
        plVar9 = *(longlong **)(param_1 + 0xf8);
        *(int **)(param_1 + 0xf8) = piVar3;
        if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))();
        }
        lVar12 = *(longlong *)(param_1 + 0x10);
        if (((*(int *)(*(longlong *)(lVar12 + 0x60) + 0x4d4) != 0) ||
            (*(int *)(lVar12 + 0x1a0) != 0)) &&
           ((iVar8 = Py_MakePendingCalls(), iVar8 < 0 && (*(longlong *)(param_1 + 0x60) != 0)))) {
          plVar15 = *(longlong **)(param_1 + 0x60);
          goto LAB_18035dbb2;
        }
        if (*(int *)(lVar12 + 0x174) != 0) {
          PyEval_SaveThread();
          PyEval_AcquireThread(param_1);
        }
        plVar9 = local_b0;
        if (*(longlong *)(param_1 + 0x80) != 0) {
          plVar9 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x80) = 0;
          uVar16 = FUN_180373490(param_1);
          *(undefined8 *)(param_1 + 0x60) = uVar16;
          if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
             (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          plVar15 = *(longlong **)(param_1 + 0x60);
          goto LAB_18035dbb2;
        }
        plVar11 = (longlong *)(**(code **)(local_b0[1] + 0xe0))(local_b0);
        pcVar17 = local_d0;
        plVar10 = local_c8;
        plVar15 = local_d8;
        local_90 = plVar11;
      } while (plVar11 != (longlong *)0x0);
    }
    plVar15 = *(longlong **)(param_1 + 0x60);
    if (plVar15 == (longlong *)0x0) {
LAB_18035dc23:
      if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
         (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
        (**(code **)(plVar9[1] + 0x30))(plVar9);
      }
      if (pcVar13 == (code *)0x0) {
        plVar9 = (longlong *)
                 PyUnicode_FromFormat
                           ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.utils.llm_retry#19="last_error" */
                            ,DAT_1803c9b98);
        plVar15 = (longlong *)
                  FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar9);
        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
        FUN_18000d1d0(param_1,plVar15);
        uVar23 = 0x55;
      }
      else {
        plVar15 = (longlong *)FUN_180373490(param_1,pcVar13);
        if (plVar15 == (longlong *)0x0) {
          plVar15 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          uVar23 = 0x55;
        }
        else {
          uVar23 = 0x55;
          FUN_18000d1d0(param_1,plVar15);
        }
      }
    }
    else {
      plVar11 = (longlong *)plVar15[1];
      if ((*(uint *)((longlong *)plVar15[1] + 0x15) & 0x40000000) == 0) {
        plVar11 = plVar15;
      }
      if ((plVar11 == *(longlong **)PyExc_StopIteration_exref) ||
         (((*(int *)(plVar11[1] + 0xa8) < 0 && ((*(uint *)(plVar11 + 0x15) & 0x40000000) != 0)) &&
          (cVar7 = FUN_180382640(), cVar7 == '\x01')))) {
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
          (**(code **)(plVar15[1] + 0x30))(plVar15);
        }
        goto LAB_18035dc23;
      }
LAB_18035dbb2:
      *(undefined8 *)(param_1 + 0x60) = 0;
      uVar23 = 0x48;
LAB_18035dbbe:
      if (((local_c8 != (longlong *)0x0) && (-1 < (int)*local_c8)) &&
         (*local_c8 = *local_c8 + -1, *local_c8 == 0)) {
        (**(code **)(local_c8[1] + 0x30))();
      }
      pcVar13 = local_d0;
      if ((-1 < (int)*local_b0) && (*local_b0 = *local_b0 + -1, *local_b0 == 0)) {
        (**(code **)(local_b0[1] + 0x30))();
        pcVar13 = local_d0;
      }
    }
  }
  piVar3 = (int *)plVar15[5];
  if (piVar3 == (int *)0x0) {
    pcVar17 = (code *)FUN_18038e2a0(local_d8,uVar23);
    plVar9 = (longlong *)plVar15[5];
    if (pcVar17 == _Py_NoneStruct_exref) {
      plVar15[5] = 0;
    }
    else {
      if ((pcVar17 != (code *)0x0) && (*(int *)pcVar17 + 1 != 0)) {
        *(int *)pcVar17 = *(int *)pcVar17 + 1;
      }
      plVar15[5] = (longlong)pcVar17;
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == local_d8) goto LAB_18035ddb7;
    pcVar17 = (code *)FUN_18038e2a0(local_d8,uVar23);
    *(int **)(pcVar17 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar9 = (longlong *)plVar15[5];
    if (pcVar17 == _Py_NoneStruct_exref) {
      pcVar14 = (code *)0x0;
    }
    else {
      pcVar14 = pcVar17;
      if ((pcVar17 != (code *)0x0) && (*(int *)pcVar17 + 1 != 0)) {
        *(int *)pcVar17 = *(int *)pcVar17 + 1;
      }
    }
    plVar15[5] = (longlong)pcVar14;
  }
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))();
  }
  if (((pcVar17 != (code *)0x0) && (-1 < *(int *)pcVar17)) &&
     (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0)) {
    (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
  }
LAB_18035ddb7:
  plVar25 = local_a8;
  plVar11 = local_c0;
  plVar10 = local_c8;
  local_110 = local_c0;
  local_118 = 0;
  local_120 = local_c8;
  local_130 = local_98;
  local_138 = local_a8;
  local_128 = pcVar13;
  FUN_18036d0a0(local_d8,"oooooooo",local_88,local_a0);
  plVar9 = DAT_1803c9af8;
  if (local_d8 == DAT_1803c9af8) {
    if ((-1 < (int)*DAT_1803c9af8) && (*DAT_1803c9af8 = *DAT_1803c9af8 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))();
    }
    DAT_1803c9af8 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
     (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1, *(longlong *)pcVar13 == 0)) {
    (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
  }
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))(plVar11);
  }
  if ((-1 < (int)*local_88) && (*local_88 = *local_88 + -1, *local_88 == 0)) {
    (**(code **)(local_88[1] + 0x30))();
  }
  if ((-1 < (int)*local_a0) && (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
    (**(code **)(local_a0[1] + 0x30))();
  }
  if ((-1 < (int)*plVar25) && (*plVar25 = *plVar25 + -1, *plVar25 == 0)) {
    (**(code **)(plVar25[1] + 0x30))(plVar25);
  }
  if ((-1 < (int)*local_98) && (*local_98 = *local_98 + -1, *local_98 == 0)) {
    (**(code **)(local_98[1] + 0x30))();
  }
  plVar9 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar15;
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))();
  }
  return 0;
}
